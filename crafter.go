package main

import (
	"fmt"
	"net/http"
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

func (crafter *Crafter) fetchItemTPPrice(itemID int) (*ItemPrice, error) {
	itemPrice, err := crafter.gw2APIClient.FetchItemPrice(itemID)
	if err != nil {
		apiErr, ok := err.(*APIError)
		if ok && apiErr.StatusCode == http.StatusNotFound {
			logger.Error("Item Price not found on TP", "itemID", itemID)
		}
		return nil, err
	}
	return itemPrice, nil
}

func (crafter *Crafter) findItemSellValue(itemID int) (int, error) {
	itemPrice, err := crafter.fetchItemTPPrice(itemID)
	if err != nil {
		return 0, err
	}
	return itemPrice.Sells.UnitPrice, nil
}

func (crafter *Crafter) findItemBuyValue(itemID int) (int, error) {
	itemPrice, err := crafter.fetchItemTPPrice(itemID)
	if err != nil {
		return 0, err
	}
	return itemPrice.Buys.UnitPrice, nil
}

func (crafter *Crafter) extractRecipeCost(recipe Recipe) (int, error) {
	ingredients := recipe.Ingredients
	recipeCost := 0
	for _, ingredient := range ingredients {
		itemPrice, err := crafter.findItemBuyValue(ingredient.ItemID)
		if err != nil {
			return 0, fmt.Errorf("failed to find item buy value: %w", err)
		}
		// TODO: Currently hardcode to use buy order price for recipe cost
		recipeCost += itemPrice * ingredient.Count
	}
	return recipeCost, nil
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

func (crafter *Crafter) calculateProfitMargin(recipe Recipe) (float64, error) {
	logger.Debug("Calculating profit margin...", "recipeID", recipe.ID, "OutputItemID", recipe.OutputItemID)
	recipeCost, err := crafter.extractRecipeCost(recipe)
	if err != nil {
		return 0, err
	}
	recipeOutputPrice, err := crafter.findItemSellValue(recipe.OutputItemID)
	if err != nil {
		return 0, err
	}
	// Assuming we are selling it on TP
	return (float64(recipeOutputPrice) * 0.85) / float64(recipeCost), nil
}

func (crafter *Crafter) FindProfitableOptions(itemID int, depth int) ([]RecipeProfit, error) {
	if depth == 0 {
		return nil, nil
	}
	if depth < 1 {
		return nil, fmt.Errorf("Cannot have depth < 1")
	}
	availableRecipes, err := crafter.localCache.GetRecipeByIngredient(itemID)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error fetching Available recipes for ItemID %d : %v\n", itemID, err))
	}
	var profitableRecipes []RecipeProfit
	for _, recipe := range availableRecipes {
		if !crafter.recipeIsViable(recipe) {
			logger.Debug("Recipe is not viable for crafting", "recipeID", recipe.ID)
			continue
		}
		profitMargin, err := crafter.calculateProfitMargin(recipe)
		if err != nil {
			return nil, err
		}
		if profitMargin < configObj.ProfitThreshold {
			logger.Debug("Recipe not profitable", "recipeID", recipe.ID, "profitMargin", profitMargin)
			continue
		}
		logger.Debug("Recipe is profitable", "recipeID", recipe.ID, "profitMargin", profitMargin)
		profitableRecipes = append(profitableRecipes, RecipeProfit{RecipeID: recipe.ID, OutputItemID: recipe.OutputItemID, ProfitMargin: profitMargin})

		subRecipes, err := crafter.FindProfitableOptions(recipe.OutputItemID, depth-1)
		if err != nil {
			logger.Fatal("Error fetching subRecipes", "itemID", recipe.OutputItemID, "PArentRecipeID", recipe.ID)
		}

		profitableRecipes = append(profitableRecipes, subRecipes...)
	}

	sort.Slice(profitableRecipes, func(i, j int) bool {
		return profitableRecipes[i].ProfitMargin > profitableRecipes[j].ProfitMargin
	})

	return profitableRecipes, nil
}
