package main

import (
	"fmt"
	"log"
	"slices"
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

func (crafter *Crafter) recipeOutputIsTradeable(itemID int) bool {
	isTradeable, err := crafter.localCache.ItemIsTradeable(itemID)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error checking if item is tradeable: %v", err))
	}
	return isTradeable
}

func (crafter *Crafter) recipeIsViable(recipe Recipe) bool {
	return crafter.recipeIsAvailable(recipe) && crafter.recipeOutputIsTradeable(recipe.OutputItemID)
}

func (crafter *Crafter) FindProfitableOptions(itemID int) (int, error) {
	availableRecipes, err := crafter.localCache.GetRecipeByIngredient(itemID)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error fetching Available recipes for ItemID %d : %v\n", itemID, err))
	}
	for _, recipe := range availableRecipes {
		if !crafter.recipeIsViable(recipe) {
			logger.Info(fmt.Sprintf("Recipe %d is not viable for crafting", recipe.ID))
			continue
		}
		logger.Info(fmt.Sprintf("Recipe Data: %+v", recipe))
	}

	return 0, nil
}
