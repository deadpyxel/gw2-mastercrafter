package main

import (
	"fmt"
	"strings"
)

type Metadata struct {
	BuildNumber int `json:"id" db:"build_number"`
}

type Ingredient struct {
	ID       int `db:"id"`
	ItemID   int `json:"item_id" db:"item_id"`
	Count    int `json:"count" db:"count"`
	RecipeID int `db:"recipe_id"`
}

type Recipe struct {
	ID              int          `json:"id" db:"id"`
	Type            string       `json:"type" db:"type"`
	OutputItemID    int          `json:"output_item_id" db:"output_item_id"`
	OutputItemCount int          `json:"output_item_count" db:"output_item_count"`
	Disciplines     StringSlice  `json:"disciplines" db:"disciplines"` // TODO: Define a StringSlice type?
	MinRating       int          `json:"min_rating" db:"min_rating"`
	Flags           StringSlice  `json:"flags" db:"flags"`
	Ingredients     []Ingredient `json:"ingredients"`
}

type Currency struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

type WalletCurrency struct {
	CurrencyID int `json:"id"`
	Value      int `json:"value"`
}

type MerchantItem struct {
	ItemID int
	Price  WalletCurrency
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
	Name        string      `json:"name" db:"name"`
	Type        string      `json:"type" db:"type"`
	Rarity      string      `json:"rarity" db:"rarity"`
	VendorValue int         `json:"vendor_value" db:"vendor_value"`
	Flags       StringSlice `json:"flags" db:"flags"`
	ID          int         `json:"id" db:"id"`
}

type StringSlice []string

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected type for StringSlice: %T", value)
	}

	*s = strings.Split(str, ",")
	return nil
}

type RecipeProfit struct {
	RecipeID     int
	ProfitMargin float64
}
