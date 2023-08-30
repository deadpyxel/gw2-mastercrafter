package main

import (
	"database/sql"
	"fmt"
	"log"

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
		return 0, err
	}
	return currentBuildNumber, err
}

func fetchAllRecipeDataFromAPI(client *APIClient) ([]Recipe, error) {
	recipesIds, err := client.FetchAllRecipesIds()
	if err != nil {
		return []Recipe{}, err
	}
	recipes := []Recipe{}
	for _, recipeId := range recipesIds {
		recipe, err := client.FetchRecipe(recipeId)
		if err != nil {
			return []Recipe{}, err
		}
		recipes = append(recipes, *recipe)
	}
	return recipes, nil
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
