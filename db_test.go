package main

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func setupDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	return sqlxDB, mock, func() {
		db.Close()
	}
}

func TestGetCurrencyIdByName(t *testing.T) {
	db, mock, cleanup := setupDB(t)
	defer cleanup()

	lc := NewLocalCache(db)
	// Define mock expectations for updateCurrencyCache
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS currencies").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT OR REPLACE INTO currencies").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	// Insert test data using the mock
	currencies := []Currency{
		{ID: 1, Name: "Coin", Description: "Currency description"},
		{ID: 2, Name: "Coin", Description: "Currency description"},
		{ID: 3, Name: "Karma", Description: "Currency description"},
	}
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
			name:           "Multiple matches",
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
