package main

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupDB(t *testing.T) (*sqlx.DB, func()) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	return db, func() {
		db.Close()
	}
}

func TestGetCurrencyIdByName(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	lc := NewLocalCache(db)

	// Insert test data using the mock
	currencies := []Currency{
		{ID: 1, Name: "Coin", Description: "Currency description"},
		{ID: 2, Name: "Coin", Description: "Currency description"},
		{ID: 3, Name: "Karma", Description: "Currency description"},
	}
	// Define mock expectations for updateCurrencyCache
	err := updateCurrencyCache(db, currencies)
	if err != nil {
		t.Fatalf("Failed to update currency cache: %v", err)
	}
	testCases := []struct {
		name           string
		currencyName   string
		expectedID     int
		expectedErrStr string
	}{
		{
			name:           "Valid currency name",
			currencyName:   "Coin",
			expectedID:     1,
			expectedErrStr: "",
		},
		{
			name:           "Invalid currency name",
			currencyName:   "Invalid",
			expectedID:     0,
			expectedErrStr: "Currency not found",
		},
		{
			name:           "Multiple matches return first match",
			currencyName:   "Coin",
			expectedID:     1,
			expectedErrStr: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := lc.GetCurrencyIDByName(tc.currencyName)
			if err != nil {
				if err.Error() != tc.expectedErrStr {
					t.Fatalf("Unexpected error: got %v, want %v", err.Error(), tc.expectedErrStr)
				}
			} else {
				if id != tc.expectedID {
					t.Fatalf("Unexpected currency id: got %v, want %v", id, tc.expectedID)
				}
			}
		})
	}
}
