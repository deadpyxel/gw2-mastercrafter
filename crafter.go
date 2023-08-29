package main

import (
	"fmt"
	"log"
)

type Crafter struct {
	gw2APIClient APIClient
}

func NewCrafter(gw2APIClient APIClient) *Crafter {
	return &Crafter{gw2APIClient: gw2APIClient}
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

func (crafter *Crafter) FindProfitableOptions(itemID int) (int, error) {
	// knownRecipes, err := crafter.gw2APIClient.FetchKnownRecipesIds()
	// if err != nil {
	// 	fmt.Printf("Error fetching Known Recipe IDs: %v\n", err)
	// }
	availableRecipes, err := crafter.gw2APIClient.FetchAvailableRecipesIds(itemID)
	if err != nil {
		fmt.Printf("Error fetching Available recipes for ItemID %d : %v\n", itemID, err)
	}
	for _, recipeId := range availableRecipes {
		// if !slices.Contains(knownRecipes, recipeId) {
		// 	fmt.Printf("Recipe %d not known, skipping...\n", recipeId)
		// 	continue
		// }
		recipe, err := crafter.gw2APIClient.FetchRecipe(recipeId)
		if err != nil {
			log.Fatalf("Recipe fetching failed for recipe ID %d: %v\n", recipeId, err)
		}
		fmt.Printf("Recipe info: %+v\n", recipe)
		recipeCost := crafter.extractRecipeCost(*recipe)
		recipeOutputPrice := crafter.findItemSellValue(recipe.OutputItemID)
		profit := float32(recipeOutputPrice) / float32(recipeCost)
		fmt.Printf("craft price: %d, sell price: %d, profit: %f\n", recipeCost, recipeOutputPrice, profit)

	}

	return 0, nil
}
