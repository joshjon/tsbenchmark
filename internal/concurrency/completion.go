package concurrency

import (
	"fmt"
	"go.uber.org/zap"
	"math"
	"sync"
	"time"
)

type CompletionFuture struct {
	*WorkerResult
	complete chan *WorkerResult
}

func newCompletionFuture(workerID int) *CompletionFuture {
	return &CompletionFuture{
		WorkerResult: newResult(workerID),
		complete:     make(chan *WorkerResult, 1),
	}
}

func (c *CompletionFuture) Get() *WorkerResult {
	return <-c.complete
}

func (c *CompletionFuture) send() {
	c.complete <- c.WorkerResult
	close(c.complete)
}

type WorkerResult struct {
	sync.RWMutex
	ID          int
	Completed   int
	Duration    time.Duration
	MinDuration time.Duration
	MaxDuration time.Duration
}

func newResult(workerID int) *WorkerResult {
	return &WorkerResult{
		ID:          workerID,
		MinDuration: time.Duration(math.Inf(1)),
		MaxDuration: time.Duration(math.Inf(-1)),
	}
}

func (b *WorkerResult) addTaskDuration(duration time.Duration) {
	b.Lock()
	defer b.Unlock()
	b.Completed++
	b.Duration += duration
	if duration < b.MinDuration {
		b.MinDuration = duration
	} else if duration > b.MaxDuration {
		b.MaxDuration = duration
	}
}

type PoolResult struct {
	TotalCompleted     int
	AggregatedDuration time.Duration
	MinTaskDuration    time.Duration
	MaxTaskDuration    time.Duration
	WorkerResults      []*WorkerResult
}

func newPoolResult(workerResults []*WorkerResult) PoolResult {
	poolResult := PoolResult{
		MinTaskDuration: time.Duration(math.Inf(1)),
		MaxTaskDuration: time.Duration(math.Inf(-1)),
		WorkerResults:   workerResults,
	}

	for _, result := range workerResults {
		poolResult.TotalCompleted += result.Completed
		poolResult.AggregatedDuration += result.Duration

		if result.MinDuration < poolResult.MinTaskDuration {
			poolResult.MinTaskDuration = result.MinDuration
		}

		if result.MaxDuration > poolResult.MaxTaskDuration {
			poolResult.MaxTaskDuration = result.MaxDuration
		}

		zap.L().Debug(fmt.Sprintf("%d - completed: %d, duration: %s", result.ID, result.Completed, result.Duration.String()))
	}

	return poolResult
}
