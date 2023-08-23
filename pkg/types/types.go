package api

type Ingredient struct {
	ItemID int `json:"item_id"`
	Count  int `json:"count"`
}

type Recipes struct {
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

type KnownRecipes []int
