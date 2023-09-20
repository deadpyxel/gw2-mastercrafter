package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func generateItemIds(size int) []int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var tradeableItemIds []int
	for i := 0; i < size; i++ {
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

	// Test performance for different batch sizes
	for _, size := range []int{100, 1000, 10000} {
		tradeableItemIds := generateItemIds(size)

		b.Run(fmt.Sprintf("BatchSize-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := updateTradeableItemsCache(db, tradeableItemIds)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
