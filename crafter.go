package main

import (
	"fmt"
	"log"
	"slices"
	"sort"
)

// A Crafter uses the API client to request data about pricing of items,
// while using the localCache to search information on crafting recipes
type Crafter struct {
	gw2APIClient APIClient  // underlying API client connection
	localCache   LocalCache // underlying local SQlite cache
}

func NewCrafter(gw2APIClient APIClient, localCache LocalCache) *Crafter {
	return &Crafter{gw2APIClient: gw2APIClient, localCache: localCache}
}

func (crafter *Crafter) fetchItemTPPrice(itemID int) *ItemPrice {
	itemPrice, err := crafter.gw2APIClient.FetchItemPrice(itemID)
	if err != nil {
		log.Fatalf("Failed to fetch item price for itemID %d: %v", itemID, err)
	}
	return itemPrice
}

func (crafter *Crafter) findItemSellValue(itemID int) int {
	itemPrice := *crafter.fetchItemTPPrice(itemID)
	return itemPrice.Sells.UnitPrice
}

func (crafter *Crafter) findItemBuyValue(itemID int) int {
	itemPrice := *crafter.fetchItemTPPrice(itemID)
	return itemPrice.Buys.UnitPrice
}

func (crafter *Crafter) extractRecipeCost(recipe Recipe) int {
	ingredients := recipe.Ingredients
	recipeCost := 0
	for _, ingredient := range ingredients {
		itemPrice := crafter.findItemBuyValue(ingredient.ItemID)
		// TODO: Currently hardcode to use buy order price for recipe cost
		recipeCost += itemPrice * ingredient.Count
	}
	return recipeCost
}

func (crafter *Crafter) recipeIsAvailable(recipe Recipe) bool {
	if slices.Contains(recipe.Flags, "AutoLearned") {
		return true
	}
	knownRecipeIds, err := crafter.gw2APIClient.FetchKnownRecipesIds()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error fetching Known Recipe IDs: %v", err))
	}
	return slices.Contains(knownRecipeIds, recipe.ID)
}

func (crafter *Crafter) itemIsTradeable(itemID int) bool {
	isTradeable, err := crafter.localCache.ItemIsTradeable(itemID)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error checking if item is tradeable: %v", err))
	}
	return isTradeable
}

func (crafter *Crafter) itemTypeisAllowed(itemType string) bool {
	return !slices.Contains(configObj.RemovedTypes, itemType)
}

// A viable recipe is
// - available (learned)
// - has an output that is tradeable
// - the output item type is not present on filtered out options
func (crafter *Crafter) recipeIsViable(recipe Recipe) bool {
	return crafter.recipeIsAvailable(recipe) && crafter.itemIsTradeable(recipe.OutputItemID) && crafter.itemTypeisAllowed(recipe.Type)
}

func (crafter *Crafter) calculateProfitMargin(recipe Recipe) float64 {
	logger.Debug("Calculating profit margin...", "recipeID", recipe.ID, "OutputItemID", recipe.OutputItemID)
	recipeCost := float64(crafter.extractRecipeCost(recipe))
	recipeOutputPrice := float64(crafter.findItemSellValue(recipe.OutputItemID))
	// Assuming we are selling it on TP
	return (recipeOutputPrice * 0.85) / recipeCost
}

func (crafter *Crafter) FindProfitableOptions(itemID int) ([]RecipeProfit, error) {
	availableRecipes, err := crafter.localCache.GetRecipeByIngredient(itemID)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error fetching Available recipes for ItemID %d : %v\n", itemID, err))
	}
	var profitableRecipes []RecipeProfit
	for _, recipe := range availableRecipes {
		if !crafter.recipeIsViable(recipe) {
			logger.Debug(fmt.Sprintf("Recipe %d is not viable for crafting", recipe.ID))
			continue
		}
		profitMargin := crafter.calculateProfitMargin(recipe)
		if profitMargin < configObj.ProfitThreshold {
			logger.Debug("Recipe not profitable", "recipeID", recipe.ID, "profitMargin", profitMargin)
			continue
		}
		logger.Debug("Recipe is profitable", "recipeID", recipe.ID, "profitMargin", profitMargin)
		profitableRecipes = append(profitableRecipes, RecipeProfit{RecipeID: recipe.ID, ProfitMargin: profitMargin})
	}

	sort.Slice(profitableRecipes, func(i, j int) bool {
		return profitableRecipes[i].ProfitMargin > profitableRecipes[j].ProfitMargin
	})

	return profitableRecipes, nil
}
