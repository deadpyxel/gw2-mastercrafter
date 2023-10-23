package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type LocalCache struct {
	db *sqlx.DB
}

func NewLocalCache(db *sqlx.DB) *LocalCache {
	return &LocalCache{db: db}
}

func (lc *LocalCache) GetRecipeById(recipeID int) (*Recipe, error) {
	var recipe Recipe
	err := lc.db.Get(&recipe, "SELECT * FROM recipes WHERE id = ?", recipeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Recipe not found")
		}
		return nil, err
	}

	var ingredients []Ingredient
	err = lc.db.Select(&ingredients, "SELECT * FROM ingredients WHERE recipe_id = ?", recipeID)
	if err != nil {
		return nil, err
	}
	recipe.Ingredients = ingredients

	return &recipe, nil
}

func (lc *LocalCache) GetRecipeByIngredient(ingredientID int) ([]Recipe, error) {
	var recipes []Recipe
	err := lc.db.Select(&recipes, `
		SELECT r.* FROM recipes r
		INNER JOIN ingredients ing ON r.id = ing.recipe_id
		WHERE ing.item_id = ?
	`, ingredientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("No recipes found with given ingredientID")
		}
		return nil, err
	}

	for i := range recipes {
		var ingredients []Ingredient
		err = lc.db.Select(&ingredients, "SELECT * FROM ingredients WHERE recipe_id = ?", recipes[i].ID)
		if err != nil {
			return nil, err
		}
		recipes[i].Ingredients = ingredients
	}

	return recipes, nil
}

func (lc *LocalCache) GetItemById(itemID int) (*Item, error) {
	var item Item
	err := lc.db.Get(&item, "SELECT * FROM items WHERE id = ?", itemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Item not found")
		}
		return nil, err
	}
	return &item, nil
}

func (lc *LocalCache) ItemIsTradeable(itemID int) (bool, error) {
	var id int
	query := "SELECT EXISTS(SELECT 1 FROM tradeable_items WHERE id = ?)"
	err := lc.db.Get(&id, query, itemID)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (lc *LocalCache) GetCurrencyIDByName(currencyName string) (int, error) {
	var id int
	err := lc.db.QueryRowx(`SELECT id FROM currencies WHERE name = ?`, currencyName).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("Currency not found")
		}
		return 0, err
	}
	return id, nil
}

func (lc *LocalCache) HasPurchaseOptionWithCurrency(itemId int, currencyName string) (bool, error) {
	currencyId, err := lc.GetCurrencyIDByName(currencyName)
	if err != nil {
		return false, err
	}
	var count int
	err = lc.db.QueryRowx(`SELECT COUNT(*) FROM purchase_options po
                           JOIN merchant_prices mp ON po.id = mp.purchase_option_id
                           WHERE po.item_id = ? AND mp.currency_id = ? AND po.ignore = 0`, itemId, currencyId).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (lc *LocalCache) GetMerchantItemPrice(itemID int, currencyName string) (*ItemPrice, error) {
	currencyId, err := lc.GetCurrencyIDByName(currencyName)
	if err != nil {
		return nil, err
	}
	var currencyPrice int
	err = lc.db.QueryRowx(
		`SELECT mp.count FROM merchant_prices merchant_prices
    JOIN purchase_options po ON mp.purchase_option_id = po.recipe_id
    WHERE po.item_id = ? AND mp.currency_id = ?`, itemID, currencyId).Scan(&currencyPrice)
	if err != nil {
		if err == sql.ErrNoRows {
			// No price was found with given currency
			return nil, fmt.Errorf("No price found in %s for itemID %d", currencyName, itemID)
		}
		// An error has occurred during the query
		return nil, err
	}
	// Price found, return a custom ItemPrice
	return &ItemPrice{
		ID:    itemID,
		Buys:  TradingPostPrice{UnitPrice: currencyPrice},
		Sells: TradingPostPrice{UnitPrice: currencyPrice},
	}, nil
}
