package concurrency

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPool_Dispatch_identicalRouteKeysOneWorkerStarted(t *testing.T) {
	maxWorkers := 1000
	wantWorkers := 1

	pool := NewPool(PoolConfig{
		MaxWorkers: maxWorkers,
	})

	pool.Dispatch()

	for i := 0; i < pool.config.MaxWorkers; i++ {
		task := &Task{
			RouteKey: "identical-route-key",
			Func: func() error {
				return nil
			},
		}
		pool.Submit(task)
		time.Sleep(time.Millisecond)

	}

	results := pool.Wait()
	assert.Len(t, pool.workers.workers, wantWorkers)
	assert.Len(t, results, wantWorkers)
}

func TestPool_Dispatch_uniqueRouteKeysMaxWorkersStarted(t *testing.T) {
	tests := []struct {
		name    string
		wantMax int
	}{
		{
			name:    "1 worker",
			wantMax: 1,
		},
		{
			name:    "10 workers",
			wantMax: 10,
		},
		{
			name:    "100 workers",
			wantMax: 100,
		},
		{
			name:    "1000 workers",
			wantMax: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool(PoolConfig{
				MaxWorkers: tt.wantMax,
			})

			pool.Dispatch()

			for i := 0; i < tt.wantMax; i++ {
				task := &Task{
					RouteKey: time.Now().String(),
					Func: func() error {
						return nil
					},
				}
				pool.Submit(task)
			}

			time.Sleep(time.Millisecond)
			results := pool.Wait()
			assert.Len(t, pool.workers.workers, tt.wantMax)
			assert.Len(t, results, tt.wantMax)
		})
	}
}
