package main

import (
	"fmt"
	"os"

	config "github.com/deadpyxel/gw2-mastercrafter/internal"
	"github.com/deadpyxel/gw2-mastercrafter/internal/api"
	"github.com/deadpyxel/gw2-mastercrafter/pkg"
	"github.com/jmoiron/sqlx"
)

var configObj config.Config
var logger *SLogLogger

// SLogLogger wrapper to implement the interface - temporary solution
type SLogLogger struct {
	log pkg.Logger
}

func (sl *SLogLogger) Debug(msg string, metadata ...interface{}) {
	// TODO: implement proper logger from internal/logger package
	fmt.Printf("DEBUG: %s %v\n", msg, metadata)
}

func (sl *SLogLogger) Info(msg string, metadata ...interface{}) {
	fmt.Printf("INFO: %s %v\n", msg, metadata)
}

func (sl *SLogLogger) Warn(msg string, metadata ...interface{}) {
	fmt.Printf("WARN: %s %v\n", msg, metadata)
}

func (sl *SLogLogger) Error(msg string, metadata ...interface{}) {
	fmt.Printf("ERROR: %s %v\n", msg, metadata)
}

func (sl *SLogLogger) Fatal(msg string, metadata ...interface{}) {
	fmt.Printf("FATAL: %s %v\n", msg, metadata)
	os.Exit(1)
}

func (sl *SLogLogger) SetLevel(level string) {
	// TODO: implement proper level setting
}

func NewSLogLogger() *SLogLogger {
	return &SLogLogger{}
}

// Temporary wrapper for LocalCache to avoid import cycles
type LocalCache struct {
	// TODO: Move to internal/cache package
}

func NewLocalCache(db *sqlx.DB) *LocalCache {
	return &LocalCache{}
}

// Temporary implementation
func (lc *LocalCache) GetRecipeById(recipeID int) (*pkg.Recipe, error)              { return nil, nil }
func (lc *LocalCache) GetRecipeByIngredient(ingredientID int) ([]pkg.Recipe, error) { return nil, nil }
func (lc *LocalCache) GetItemById(itemID int) (*pkg.Item, error)                    { return nil, nil }
func (lc *LocalCache) ItemIsTradeable(itemID int) (bool, error)                     { return false, nil }
func (lc *LocalCache) GetCurrencyIDByName(currencyName string) (int, error)         { return 0, nil }
func (lc *LocalCache) HasPurchaseOptionWithCurrency(itemId int, currencyName string) (bool, error) {
	return false, nil
}
func (lc *LocalCache) GetMerchantItemPrice(itemID int, currencyName string) (*pkg.ItemPrice, error) {
	return nil, nil
}

// Temporary Crafter implementation
type Crafter struct{}

func NewCrafter(apiClient pkg.GW2APIClient, cache pkg.Cache, logger pkg.Logger, config pkg.Config) *Crafter {
	return &Crafter{}
}

func (c *Crafter) FindProfitableOptions(itemID int, depth int) ([]pkg.RecipeProfit, error) {
	return nil, nil
}

// Temporary UpdateCache function
func UpdateCache(client pkg.GW2APIClient) {
	fmt.Println("Cache update functionality temporarily disabled during refactoring")
}

func main() {
	// Load API Token, create API client instance
	apiToken := os.Getenv("API_TOKEN")
	configObj = config.ReadConfig()
	logger = NewSLogLogger()
	logger.SetLevel(configObj.LogLevel)
	if apiToken == "" {
		apiToken = configObj.ApiKey
	}
	gw2Client := api.NewAPIClient("https://api.guildwars2.com/v2", apiToken)

	UpdateCache(gw2Client)

	// Initialize Local SQLite Cache connection
	db, err := sqlx.Connect("sqlite3", "cache.db")
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error updating local cache: %v", err))
	}
	defer db.Close()
	localCache := NewLocalCache(db)

	// Create crafter instance
	crafter := NewCrafter(gw2Client, localCache, logger, &configObj)
	targetItems := []int{19718, 19739, 19741, 19743, 19748, 19745, 19719, 19728, 19730, 19731, 19729, 19732, 19697, 19704, 19703, 19699, 19698, 19702, 19700, 19701, 19723, 19726, 19727, 19724, 19722, 19725}
	for _, targetItem := range targetItems {
		profitableRecipes, err := crafter.FindProfitableOptions(targetItem, 1)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Error finding profitable options: %s", err.Error()), "itemID", targetItem)
		}
		logger.Info(fmt.Sprintf("Found %d profitable recipes for itemID %d: %v", len(profitableRecipes), targetItem, profitableRecipes), "itemID", targetItem)
	}
}
