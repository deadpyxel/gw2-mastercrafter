package main

import (
	"time"
)

// BatchFetchConfig holds configuration for batch operations
type BatchFetchConfig struct {
	Concurrency int
	BatchSize   int
	RateLimit   time.Duration
	Retries     int
	BaseDelay   time.Duration
}

// DefaultBatchFetchConfig returns default configuration for GW2 API
func DefaultBatchFetchConfig() BatchFetchConfig {
	return BatchFetchConfig{
		Concurrency: 8,
		BatchSize:   200,
		RateLimit:   time.Minute / 300, // GW2 API has 300 requests/minute rate limit
		Retries:     3,
		BaseDelay:   time.Second,
	}
}

// BatchFetcher provides generic batch fetching functionality
type BatchFetcher[T any, IDType any] struct {
	config BatchFetchConfig
	logger Logger
}

// NewBatchFetcher creates a new batch fetcher with the given configuration
func NewBatchFetcher[T any, IDType any](config BatchFetchConfig, logger Logger) *BatchFetcher[T, IDType] {
	return &BatchFetcher[T, IDType]{
		config: config,
		logger: logger,
	}
}

// FetchBatch represents a function that fetches a batch of items given a slice of IDs
type FetchBatch[T any, IDType any] func(ids []IDType) ([]T, error)

// FetchAll executes batch fetching with concurrency, rate limiting, and retry logic
func (bf *BatchFetcher[T, IDType]) FetchAll(
	allIds []IDType,
	fetchBatch FetchBatch[T, IDType],
	debugName string,
) ([]T, error) {
	bf.logger.Debug("Starting batch fetch", "type", debugName, "totalItems", len(allIds))

	resultChannel := make(chan []T)
	idsChannel := make(chan []IDType)
	errorChannel := make(chan error)
	doneChannel := make(chan struct{})

	ticker := time.NewTicker(bf.config.RateLimit)
	defer ticker.Stop()

	// Start worker goroutines
	for i := 0; i < bf.config.Concurrency; i++ {
		go func() {
			for idsBatch := range idsChannel {
				var result []T
				var err error
				retries := bf.config.Retries
				delay := bf.config.BaseDelay

				for retries > 0 {
					<-ticker.C // Rate limiting
					result, err = fetchBatch(idsBatch)
					if err == nil {
						break
					}
					if isRetriable(err) {
						time.Sleep(delay)
						delay *= 2
						retries--
						continue
					}
					// Non-retriable error, break out
					break
				}
				if err != nil {
					errorChannel <- err
					return
				}
				resultChannel <- result
			}
			doneChannel <- struct{}{}
		}()
	}

	// Distribute work
	go func() {
		for i := 0; i < len(allIds); i += bf.config.BatchSize {
			end := i + bf.config.BatchSize
			if end > len(allIds) {
				end = len(allIds)
			}
			idsChannel <- allIds[i:end]
		}
		close(idsChannel)
	}()

	// Collect results
	var allResults []T
	completedGoroutines := 0
	for completedGoroutines < bf.config.Concurrency {
		select {
		case batch := <-resultChannel:
			allResults = append(allResults, batch...)
		case err := <-errorChannel:
			return nil, err
		case <-doneChannel:
			completedGoroutines++
		}
	}

	bf.logger.Debug("Completed batch fetch", "type", debugName, "fetchedItems", len(allResults))
	return allResults, nil
}
