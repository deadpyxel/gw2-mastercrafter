package main

import (
	"fmt"
	"os"

	config "github.com/deadpyxel/gw2-mastercrafter/internal"
	"github.com/jmoiron/sqlx"
)

var configObj config.Config

func main() {
	// Load API Token, create API client instance
	apiToken := os.Getenv("API_TOKEN")
	configObj = config.ReadConfig()
	if apiToken == "" {
		apiToken = configObj.ApiKey
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
	targetItems := []int{19718, 19739, 19741, 19743, 19748, 19745, 19719, 19728, 19730, 19731, 19729, 19732, 19697, 19704, 19703, 19699, 19698, 19702, 19700, 19701, 19723, 19726, 19727, 19724, 19722, 19725}
	for _, targetItem := range targetItems {
		profitableRecipes, err := crafter.FindProfitableOptions(targetItem)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Error finding profitable options: %s", err.Error()), "itemID", targetItem)
		}
		logger.Info(fmt.Sprintf("Found %d profitable recipes for itemID %d: %v", len(profitableRecipes), targetItem, profitableRecipes), "itemID", targetItem)
	}
}
