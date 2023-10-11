package main

import (
	"encoding/json"
	"os"
)

type MerchantPrice struct {
	Type  string `json:"type"`  // type of currency for payment (Item when it is an exchange)
	ID    int    `json:"id"`    // item/currency id
	Count int    `json:"count"` // count of item/currency to use
}

type MerchantOptions struct {
	Type   string          `json:"type"`             // type of item offering
	ID     int             `json:"id"`               // Item id
	Count  int             `json:"count"`            // count of item for sale
	Price  []MerchantPrice `json:"price"`            // pricing options for the item
	Ignore bool            `json:"ignore,omitempty"` // True if item should be ignored for processing (deprecated or unavailable)
}

type Merchant struct {
	Name            string            `json:"name" db:"name"`                           // Merchant name
	Locations       []string          `json:"locations" db:"location"`                  // Merchant locations
	DisplayName     string            `json:"display_name,omitempty" db:"display_name"` // Display name for merchant
	PurchaseOptions []MerchantOptions `json:"purchase_options"`                         // Offerings by merchant
}

func ParseMerchantDataFile(filepath string) ([]Merchant, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var merchants []Merchant
	err = json.Unmarshal(data, &merchants)
	if err != nil {
		return nil, err
	}
	return merchants, nil
}
