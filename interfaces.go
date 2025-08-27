package main

// GW2APIClient defines the interface for interacting with Guild Wars 2 API
type GW2APIClient interface {
	// Item and pricing operations
	FetchItemPrice(itemID int) (*ItemPrice, error)
	FetchItem(itemID int) (*Item, error)
	BatchFetchItems(itemIds []int) ([]Item, error)
	FetchAllItemsIds() ([]int, error)

	// Recipe operations
	FetchRecipe(recipeID int) (*Recipe, error)
	BatchFetchRecipes(recipeIds RecipeIds) ([]Recipe, error)
	FetchAvailableRecipesIds(itemID int) (RecipeIds, error)
	FetchKnownRecipesIds() (RecipeIds, error)
	FetchAllRecipesIds() (RecipeIds, error)

	// Metadata and currencies
	FetchBuildNumber() (Metadata, error)
	FetchCurrencies() ([]Currency, error)
	FetchAllIds(endpoint string) ([]int, error)

	// Generic batch operations
	BatchFetch(ids []int, endpoint string, dataType string) ([]interface{}, error)
}

// Cache defines the interface for local data storage operations
type Cache interface {
	// Recipe operations
	GetRecipeById(recipeID int) (*Recipe, error)
	GetRecipeByIngredient(ingredientID int) ([]Recipe, error)

	// Item operations
	GetItemById(itemID int) (*Item, error)
	ItemIsTradeable(itemID int) (bool, error)

	// Currency and merchant operations
	GetCurrencyIDByName(currencyName string) (int, error)
	HasPurchaseOptionWithCurrency(itemId int, currencyName string) (bool, error)
	GetMerchantItemPrice(itemID int, currencyName string) (*ItemPrice, error)
}

// Logger defines the interface for logging operations
type Logger interface {
	Debug(msg string, metadata ...interface{})
	Info(msg string, metadata ...interface{})
	Warn(msg string, metadata ...interface{})
	Error(msg string, metadata ...interface{})
	Fatal(msg string, metadata ...interface{})
	SetLevel(level string)
}

// Config defines the interface for configuration access
type Config interface {
	GetProfitThreshold() float64
	GetRemovedTypes() []string
}
