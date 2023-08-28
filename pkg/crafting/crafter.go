package crafting

import (
	"fmt"
	"log"
	"slices"

	"github.com/deadpyxel/gw2-mastercrafter/pkg/api"
	"github.com/deadpyxel/gw2-mastercrafter/pkg/types"
)

type Crafter struct {
	gw2APIClient api.APIClient
}

func NewCrafter(gw2APIClient api.APIClient) *Crafter {
	return &Crafter{gw2APIClient: gw2APIClient}
}

func extractRecipeCost(recipe types.Recipe) int {
	ingredients := recipe.Ingredients
	recipeCost := 0
	for _, ingredient := range ingredients {
		fmt.Printf("ingr: %v\n", ingredient)
	}
	return recipeCost
}

func (crafter *Crafter) FindProfitableOptions(itemID int) (int, error) {
	knownRecipes, err := crafter.gw2APIClient.FetchKnownRecipesIds()
	if err != nil {
		fmt.Printf("Error fetching Known Recipe IDs: %v\n", err)
	}
	availableRecipes, err := crafter.gw2APIClient.FetchAvailableRecipesIds(itemID)
	if err != nil {
		fmt.Printf("Error fetching Available recipes for ItemID %d : %v\n", itemID, err)
	}
	for _, recipeId := range availableRecipes {
		if !slices.Contains(knownRecipes, recipeId) {
			fmt.Printf("Recipe %d not known, skipping...\n", recipeId)
			continue
		}
		recipe, err := crafter.gw2APIClient.FetchRecipe(recipeId)
		if err != nil {
			log.Fatalf("Recipe fetching failed for recipe ID %d: %v\n", recipeId, err)
		}
		extractRecipeCost(*recipe)
	}

	return 0, nil
}
