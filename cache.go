package main

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func UpdateCache(client *APIClient) {
	db, err := sqlx.Connect("sqlite3", "cache.db")
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error updating local cache: %v", err))
	}
	defer db.Close()

	currentBuildMetadata, err := client.FetchBuildNumber()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Cannot fetch build number from API: %v", err))
	}
	storedBuildNumber, err := fetchStoredBuildNumber(db)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Cannot fetch stored build number from local cache: %v", err))
	}
	if currentBuildMetadata.BuildNumber > storedBuildNumber {
		// update cache
		logger.Info(fmt.Sprintf("Found new build %d, updating local cache...", currentBuildMetadata.BuildNumber))
		recipes, err := fetchAllRecipeDataFromAPI(client)
		if err != nil {
			logger.Fatal(fmt.Sprintf("%v", err))
		}
		err = updateRecipeCache(db, recipes)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failure updating local recipe cache: %+v", err))
		}
		items, err := fetchAllItemDataFromAPI(client)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failed to fetch item data from API: %v", err))
		}
		err = updateItemCache(db, items)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failure updating local item cache: %v", err))
		}
		tradeableItemIds, err := client.FetchAllIds("/commerce/prices")
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failed to fetch tradeable item ids from API: %v", err))
		}
		err = updateTradeableItemsCache(db, tradeableItemIds)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failure updating local tradeable item cache: %v", err))
		}
		// update build number here
		err = updateBuildMetadata(db, currentBuildMetadata)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Could not update Build metadata: %+v", err))
		}
	}
}

func LoadCache() []Recipe {
	db, err := sqlx.Connect("sqlite3", "cache.db")
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error updating cache: %v", err))
	}
	defer db.Close()

	recipes, err := loadRecipeCache(db)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error loading recipe cache: %v\n", err))
	}

	return recipes
}

func fetchStoredBuildNumber(db *sqlx.DB) (int, error) {
	// Check if table already exists
	var exists bool
	err := db.Get(&exists, "SELECT EXISTS (SELECT name FROM sqlite_schema WHERE type='table' AND name='metadata')")
	if err != nil {
		return 0, fmt.Errorf("failed to check if table exists: %w", err)
	}
	if !exists {
		// The table does not exist
		return 0, nil
	}
	query := "SELECT build_number FROM metadata"
	var currentBuildNumber int
	err = db.QueryRow(query).Scan(&currentBuildNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			// The table is empty
			return 0, nil
		}
		return 0, fmt.Errorf("Failed to query local cache for build number: %w", err)
	}
	return currentBuildNumber, err
}

func fetchAllRecipeDataFromAPI(client *APIClient) ([]Recipe, error) {
	recipesIds, err := client.FetchAllRecipesIds()
	if err != nil {
		return []Recipe{}, err
	}

	recipeChannel := make(chan []Recipe)
	recipeIdsChannel := make(chan RecipeIds)
	errorChannel := make(chan error)
	doneChannel := make(chan struct{})

	concurrency := 8
	batchSize := 200
	ticker := time.NewTicker(time.Minute / 300) // GW2 API has 300 requests/minute rate limit

	// start goroutines
	for i := 0; i < concurrency; i++ {
		go func() {
			for recipeIdsBatch := range recipeIdsChannel {
				var recipes []Recipe
				var err error
				retries := 3

				for retries > 0 {
					<-ticker.C
					recipes, err = client.BatchFetchRecipes(recipeIdsBatch)
					if err == nil || isRetriable(err) {
						break
					}
					retries--
				}
				if err != nil {
					errorChannel <- err
					return
				}
				recipeChannel <- recipes
			}
			doneChannel <- struct{}{}
		}()
	}

	// distribute work
	go func() {
		for i := 0; i < len(recipesIds); i += batchSize {
			end := i + batchSize
			if end > len(recipesIds) {
				end = len(recipesIds)
			}
			recipeIdsChannel <- recipesIds[i:end]
		}
		close(recipeIdsChannel)
	}()

	recipes := []Recipe{}
	completedGoroutines := 0
	for completedGoroutines < concurrency {
		select {
		case recipesBatch := <-recipeChannel:
			recipes = append(recipes, recipesBatch...)
		case err := <-errorChannel:
			return []Recipe{}, err
		case <-doneChannel:
			completedGoroutines++
		}
	}
	return recipes, nil
}
func fetchAllItemDataFromAPI(client *APIClient) ([]Item, error) {
	itemIds, err := client.FetchAllItemsIds()
	if err != nil {
		return []Item{}, err
	}

	itemChannel := make(chan []Item)
	itemIdsChannel := make(chan []int)
	errorChannel := make(chan error)
	doneChannel := make(chan struct{})

	concurrency := 8
	batchSize := 200
	ticker := time.NewTicker(time.Minute / 300) // GW2 API has 300 requests/minute rate limit

	// start goroutines
	for i := 0; i < concurrency; i++ {
		go func() {
			for itemIdsBatch := range itemIdsChannel {
				var items []Item
				var err error
				retries := 3

				for retries > 0 {
					<-ticker.C
					items, err = client.BatchFetchItems(itemIdsBatch)
					if err == nil || isRetriable(err) {
						break
					}
					retries--
				}
				if err != nil {
					errorChannel <- err
					return
				}
				itemChannel <- items
			}
			doneChannel <- struct{}{}
		}()
	}

	// distribute work
	go func() {
		for i := 0; i < len(itemIds); i += batchSize {
			end := i + batchSize
			if end > len(itemIds) {
				end = len(itemIds)
			}
			itemIdsChannel <- itemIds[i:end]
		}
		close(itemIdsChannel)
	}()

	items := []Item{}
	completedGoroutines := 0
	for completedGoroutines < concurrency {
		select {
		case itemBatch := <-itemChannel:
			items = append(items, itemBatch...)
		case err := <-errorChannel:
			return []Item{}, err
		case <-doneChannel:
			completedGoroutines++
		}
	}
	return items, nil
}

func updateMerchantItems(db *sqlx.DB, merchantItems []MerchantItem) error {
	return nil
}

func updateBuildMetadata(db *sqlx.DB, buildMetadata Metadata) error {
	if err := createMetadataTableIfNotExists(db); err != nil {
		return err
	}

	query := "UPDATE metadata set build_number = ?"
	_, err := db.Exec(query, buildMetadata.BuildNumber)
	return err
}

func updateTradeableItemsCache(db *sqlx.DB, tradeableItemIds []int) error {
	logger.Debug("Updating local tradeable items cache")
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS tradeable_items (
			id INTEGER PRIMARY KEY
		);
    CREATE INDEX IF NOT EXISTS idx_tradeable_items_id ON tradeable_items (id);
  `
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	stmt, err := tx.Preparex("INSERT OR REPLACE INTO tradeable_items (id) VALUES (?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range tradeableItemIds {
		_, err = stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func updateRecipeCache(db *sqlx.DB, recipes []Recipe) error {
	logger.Debug("Updating local Recipe cache")
	createtableQuery := `
    CREATE TABLE IF NOT EXISTS recipes (
      id INTEGER PRIMARY KEY,
      type TEXT,
      output_item_id INTEGER,
      output_item_count INTEGER,
      disciplines TEXT,
      min_rating INTEGER,
      flags TEXT
     );

     CREATE TABLE IF NOT EXISTS ingredients (
      id INTEGER PRIMARY KEY,
      item_id INTEGER,
      count INTEGER,
      recipe_id INTEGER,
      FOREIGN KEY(recipe_id) REFERENCES recipes(id),
      UNIQUE(item_id, recipe_id)
    );
  `
	_, err := db.Exec(createtableQuery)
	if err != nil {
		return err
	}
	// create transation object
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// Use a waitGroup to ensure all upserts are done before commiting
	var wg sync.WaitGroup
	errorsChan := make(chan error, len(recipes))
	upsertRecipeStmt := `
		INSERT INTO recipes (id, type, output_item_id, output_item_count, disciplines, min_rating, flags)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			output_item_id = excluded.output_item_id,
			output_item_count = excluded.output_item_count,
			disciplines = excluded.disciplines,
			min_rating = excluded.min_rating,
			flags = excluded.flags;
  `
	upsertIngredientStmt := `
		INSERT INTO ingredients (item_id, count, recipe_id)
		VALUES (?, ?, ?)
		ON CONFLICT(item_id, recipe_id) DO UPDATE SET
			count = excluded.count
  `
	for _, recipe := range recipes {
		wg.Add(1)
		go func(recipe Recipe) {
			defer wg.Done()

			_, err := tx.Exec(upsertRecipeStmt,
				recipe.ID,
				recipe.Type,
				recipe.OutputItemID,
				recipe.OutputItemCount,
				strings.Join(recipe.Disciplines, ","),
				recipe.MinRating,
				strings.Join(recipe.Flags, ","),
			)
			if err != nil {
				errorsChan <- fmt.Errorf("Error upserting recipe %d: %w", recipe.ID, err)
				return
			}

			for _, ingredient := range recipe.Ingredients {
				_, err := tx.Exec(upsertIngredientStmt, ingredient.ItemID, ingredient.Count, recipe.ID)
				if err != nil {
					errorsChan <- fmt.Errorf("Error upserting ingredient %d for recipe %d: %w", ingredient.ItemID, recipe.ID, err)
					return
				}
			}
		}(recipe)
	}

	// Wait for all upserts to finish
	wg.Wait()
	close(errorsChan)
	// Check if we had any errors during upsert
	for err := range errorsChan {
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Failed to commit transiction while updating recipe cache: %w", err)
	}
	return nil
}

func updateItemCache(db *sqlx.DB, items []Item) error {
	logger.Debug("Updating local Item cache...")
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS items (
      id INTEGER PRIMARY KEY,
      name TEXT,
      type TEXT,
      rarity TEXT,
      vendor_value INTEGER,
      flags TEXT
    );
  `
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return err
	}

	// create transaction object
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// Use a waitGroup to ensure all upserts are done before committing
	var wg sync.WaitGroup
	errorsChan := make(chan error, len(items))
	upsertItemStmt := `
		INSERT INTO items (id, name, type, rarity, vendor_value, flags)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			type = excluded.type,
			rarity = excluded.rarity,
			vendor_value = excluded.vendor_value,
			flags = excluded.flags;
  `

	for _, item := range items {
		wg.Add(1)
		go func(item Item) {
			defer wg.Done()

			_, err := tx.Exec(upsertItemStmt,
				item.ID,
				item.Name,
				item.Type,
				item.Rarity,
				item.VendorValue,
				strings.Join(item.Flags, ","),
			)
			if err != nil {
				errorsChan <- fmt.Errorf("Error upserting item %d: %w", item.ID, err)
				return
			}
		}(item)
	}

	// Wait for all upserts to finish
	wg.Wait()
	close(errorsChan)

	// Check if we had any errors during upsert
	for err := range errorsChan {
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Failed to commit transaction while updating item cache: %w", err)
	}
	return nil
}

func loadRecipeCache(db *sqlx.DB) ([]Recipe, error) {
	rows, err := db.Queryx("SELECT id, type, output_item_id, output_item_count, disciplines, min_rating, flags FROM recipes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Load recipe data on slice
	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
		var disciplines, flags string
		if err := rows.Scan(&recipe.ID, &recipe.Type, &recipe.OutputItemID, &recipe.OutputItemCount, &disciplines, &recipe.MinRating, &flags); err != nil {
			return nil, err
		}
		recipe.Disciplines = strings.Split(disciplines, ",")
		recipe.Flags = strings.Split(flags, ",")
		recipes = append(recipes, recipe)
	}

	// Fetch Ingredients
	rows, err = db.Queryx("SELECT item_id, count, recipe_id FROM ingredients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// stuff like stuff
	ingredients := make(map[int][]Ingredient)
	for rows.Next() {
		var ingredient Ingredient
		var recipeID int
		if err := rows.Scan(&ingredient.ItemID, &ingredient.Count, &recipeID); err != nil {
			return nil, err
		}
		ingredients[recipeID] = append(ingredients[recipeID], ingredient)
	}

	for i, recipe := range recipes {
		if ing, ok := ingredients[recipe.ID]; ok {
			recipes[i].Ingredients = ing
		}
	}

	return recipes, nil
}

func createMetadataTableIfNotExists(db *sqlx.DB) error {
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS metadata (
			build_number INTEGER PRIMARY KEY
		)
  `
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return err
	}
	// Check if the metadata table is empty
	checkEmptyQuery := "SELECT COUNT(*) FROM metadata"
	var count int
	err = db.Get(&count, checkEmptyQuery)
	if err != nil {
		return err
	}

	// If the table is empty, insert initial data structure
	if count == 0 {
		insertInitialDataQuery := "INSERT INTO metadata (build_number) VALUES (?)"
		_, err := db.Exec(insertInitialDataQuery, 0) // Set initial build number
		if err != nil {
			return err
		}
	}

	return nil
}
