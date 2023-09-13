package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func UpdateCache(client *APIClient) {
	db, err := sql.Open("sqlite3", "cache.db")
	if err != nil {
		log.Fatalf("Error updating cache: %v", err)
	}
	defer db.Close()

	currentBuildMetadata, err := client.FetchBuildNumber()
	if err != nil {
		log.Fatalf("Cannot fetch build number from API: %v", err)
	}
	storedBuildNumber, err := fetchStoredBuildNumber(db)
	if err != nil {
		log.Fatalf("Cannot fetch stored build number from local cache: %v", err)
	}
	if currentBuildMetadata.BuildNumber > storedBuildNumber {
		// update cache
		fmt.Printf("Found new build %d, updating local cache...\n", currentBuildMetadata.BuildNumber)
		recipes, err := fetchAllRecipeDataFromAPI(client)
		if err != nil {
			log.Fatal(err)
		}
		err = updateRecipeCache(db, recipes)
		if err != nil {
			log.Fatalf("Failure updating local recipe cache: %+v", err)
		}
		// update build number here
		err = updateBuildMetadata(db, currentBuildMetadata)
		if err != nil {
			log.Fatalf("Could not update Build metadata: %+v", err)
		}
	}
}

func LoadCache() []Recipe {
	db, err := sql.Open("sqlite3", "cache.db")
	if err != nil {
		log.Fatalf("Error updating cache: %v", err)
	}
	defer db.Close()

	recipes, err := loadRecipeCache(db)
	if err != nil {
		log.Fatalf("Error loading recipe cache: %v\n", err)
	}

	return recipes
}

func fetchStoredBuildNumber(db *sql.DB) (int, error) {
	// Check if table already exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT name FROM sqlite_schema WHERE type='table' AND name='metadata')").Scan(&exists)
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

func isRetriable(httpError error) bool {
	fmt.Printf("%+v", httpError)
	return true
}

func updateBuildMetadata(db *sql.DB, buildMetadata Metadata) error {
	if err := createMetadataTableIfNotExists(db); err != nil {
		return err
	}

	query := "UPDATE metadata set build_number = ?"
	_, err := db.Exec(query, buildMetadata.BuildNumber)
	return err
}

func updateRecipeCache(db *sql.DB, recipes []Recipe) error {
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
	tx, err := db.Begin()
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

func loadRecipeCache(db *sql.DB) ([]Recipe, error) {
	rows, err := db.Query("SELECT id, type, output_item_id, output_item_count, disciplines, min_rating, flags FROM recipes")
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
	rows, err = db.Query("SELECT item_id, count, recipe_id FROM ingredients")
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

func createMetadataTableIfNotExists(db *sql.DB) error {
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
	err = db.QueryRow(checkEmptyQuery).Scan(&count)
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
