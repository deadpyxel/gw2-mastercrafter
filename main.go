package main

import (
	"fmt"
	"log"
	"os"

	"github.com/deadpyxel/gw2-mastercrafter/internal/config"
	"github.com/deadpyxel/gw2-mastercrafter/pkg/api"
)

func main() {
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		config := config.ReadConfig()
		apiToken = config.ApiKey
	}
	gw2Client := api.NewAPIClient("https://api.guildwars2.com/v2", apiToken)
	recipeIds, err := gw2Client.FetchAvailableRecipesIds(46731)
	if err != nil {
		log.Fatalf("Error fetching Recipe Ids for process: %v\n", err)
	}
	fmt.Printf("recipeIds: %v\n", recipeIds)
}
