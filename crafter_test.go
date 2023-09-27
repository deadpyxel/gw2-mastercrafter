package main

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestFindProfitableOptions(t *testing.T) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	apiClient := NewAPIClient("localhost", "token")
	localCache := NewLocalCache(db)

	crafter := NewCrafter(*apiClient, *localCache)

	tests := []struct {
		name    string
		itemID  int
		depth   int
		wantErr bool
		want    []RecipeProfit
	}{
		{name: "Depth 0 returns nothing", itemID: 0, depth: 0, wantErr: false, want: nil},
		{name: "Negative Depth returns error", itemID: 0, depth: -1, wantErr: true, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profitableOptions, err := crafter.FindProfitableOptions(tt.itemID, tt.depth)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindProfitableOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if profitableOptions != nil {
				t.Errorf("FindProfitableOptions() returned %v when nil was expected", profitableOptions)
			}
		})
	}

}
