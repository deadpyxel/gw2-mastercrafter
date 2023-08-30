package main

import "fmt"

type Metadata struct {
	BuildNumber int `json:"id"`
}

type Ingredient struct {
	ItemID int `json:"item_id"`
	Count  int `json:"count"`
}

type Recipe struct {
	ID              int          `json:"id"`
	Type            string       `json:"type"`
	OutputItemID    int          `json:"output_item_id"`
	OutputItemCount int          `json:"output_item_count"`
	Disciplines     []string     `json:"disciplines"`
	MinRating       int          `json:"min_rating"`
	Flags           []string     `json:"flags"`
	Ingredients     []Ingredient `json:"ingredients"`
}

type TradingPostPrice struct {
	Quantity  int `json:"quantity"`
	UnitPrice int `json:"unit_price"`
}

type ItemPrice struct {
	ID    int              `json:"id"`
	Buys  TradingPostPrice `json:"buys"`
	Sells TradingPostPrice `json:"sells"`
}

func (tpPrice TradingPostPrice) String() string {
	goldAmount := tpPrice.UnitPrice / 10000
	silverAmount := (tpPrice.UnitPrice % 10000) / 100
	copperAmount := tpPrice.UnitPrice % 100
	priceString := fmt.Sprintf("%dg %ds %dc", goldAmount, silverAmount, copperAmount)
	return fmt.Sprintf("Price: %s, Orders: %d", priceString, tpPrice.Quantity)
}

type RecipeIds []int

type Item struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Rarity      string   `json:"rarity"`
	VendorValue int      `json:"vendor_value"`
	Flags       []string `json:"flags"`
	ID          int      `json:"id"`
}
