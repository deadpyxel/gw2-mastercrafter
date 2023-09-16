package main

import (
	"fmt"
	"os"

	config "github.com/deadpyxel/gw2-mastercrafter/internal"
	"github.com/jmoiron/sqlx"
)

func main() {
	// Load API Token, create API client instance
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		config := config.ReadConfig()
		apiToken = config.ApiKey
	}
	gw2Client := NewAPIClient("https://api.guildwars2.com/v2", apiToken)

	UpdateCache(gw2Client)

	// Initialize Local SQLite Cache connection
	db, err := sqlx.Connect("sqlite3", "cache.db")
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error updating local cache: %v", err))
	}
	defer db.Close()
	localCache := NewLocalCache(db)

	// Create crafter instance
	crafter := NewCrafter(*gw2Client, *localCache)
	crafter.FindProfitableOptions(19700)
}
