package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func generateItemIds() []int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var tradeableItemIds []int
	for i := 0; i < 1000; i++ {
		tradeableItemIds = append(tradeableItemIds, r.Intn(10000))
	}

	return tradeableItemIds
}

func BenchmarkUpdateTradeableItemsCache(b *testing.B) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	tradeableItemIds := generateItemIds()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := updateTradeableItemsCache(db, tradeableItemIds)
		if err != nil {
			b.Fatal(err)
		}
	}
}
