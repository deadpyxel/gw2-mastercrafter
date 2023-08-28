package main

import (
	"fmt"
	"log"
	"os"

	"github.com/deadpyxel/gw2-mastercrafter/internal/config"
	"github.com/deadpyxel/gw2-mastercrafter/pkg/api"
	"github.com/deadpyxel/gw2-mastercrafter/pkg/crafting"
)

func main() {
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		config := config.ReadConfig()
		apiToken = config.ApiKey
	}
	gw2Client := api.NewAPIClient("https://api.guildwars2.com/v2", apiToken)
	recipeIds, err := gw2Client.FetchAvailableRecipesIds(19700)
	if err != nil {
		log.Fatalf("Error fetching Recipe Ids for process: %v\n", err)
	}
	for _, recipeId := range recipeIds {
		recipe, err := gw2Client.FetchRecipe(recipeId)
		if err != nil {
			log.Fatalf("Error fetching recipe data %v\n", err)
		}
		fmt.Printf("Recipe: %v\n", recipe)
		item, err := gw2Client.FetchItem(recipe.OutputItemID)
		if err != nil {
			log.Fatalf("Error fetching item data %v\n", err)
		}
		fmt.Printf("Item Data: %v\n", item)

	}
	crafter := crafting.NewCrafter(*gw2Client)
	crafter.FindProfitableOptions(19700)
}
