package main

import (
	"database/sql"
	"fmt"
	"log"
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
		recipes, err := fetchAllRecipeDataFromAPI(client)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Recipe count: %d\n", len(recipes))
		// update build number here
		err = updateBuildMetadata(db, currentBuildMetadata)
		if err != nil {
			log.Fatalf("Could not update Build metadata: %+v", err)
		}
	}
}

func fetchStoredBuildNumber(db *sql.DB) (int, error) {
	query := "SELECT build_number FROM metadata"
	var currentBuildNumber int
	err := db.QueryRow(query).Scan(&currentBuildNumber)
	if err != nil {
		// TODO: better logic for cases where the table is empty or does not exist
		return 0, nil // in case the table does not exist
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
