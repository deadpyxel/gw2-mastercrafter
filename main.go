package main

import (
	"fmt"
	"os"

	config "github.com/deadpyxel/gw2-mastercrafter/internal"
)

func main() {
	// load stuff, check build number, update cache if needed
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		config := config.ReadConfig()
		apiToken = config.ApiKey
	}
	gw2Client := NewAPIClient("https://api.guildwars2.com/v2", apiToken)
	UpdateCache(gw2Client)
	recipes := LoadCache()
	fmt.Printf("Recipe count: %d, Recipe slice: %v\n", len(recipes), recipes[:10])
	crafter := NewCrafter(*gw2Client)
	crafter.FindProfitableOptions(19700)
}
