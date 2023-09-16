package main

import (
	"database/sql"
	"errors"

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
