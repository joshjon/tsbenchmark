package main

import (
	"fmt"
	"github.com/joshjon/tsbenchmark/internal/concurrency"
	"github.com/joshjon/tsbenchmark/internal/reader"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"time"
)

const (
	defaultMaxWorkers      = 1
	defaultWorkerQueueSize = 1000
	defaultWaitQueueSize   = 1000
	defaultDebug           = false
)

var config Config

type Config struct {
	MaxWorkers      int
	WorkerQueueSize int
	WaitQueueSize   int
	Debug           bool
}

// TODO:
//  Spin up timescale instance
//  Read from file and construct query functions (tasks)
//  Finish benchmarks
//  Handle timeouts
//  Use read only and write only channels where applicable
func main() {
	cmd := &cobra.Command{
		Use: "tsbenchmark csv_file",
		Long: "tsbenchmark is used to benchmark select query performance across " +
			"multiple workers/clients against a timescale database",
		RunE: run,
		Args: func(cmd *cobra.Command, args []string) error {
			return cobra.ExactArgs(1)(cmd, args)
		},
	}

	cmd.Flags().IntVarP(&config.MaxWorkers, "max-workers", "m", defaultMaxWorkers, "max number of concurrent workers")
	cmd.Flags().IntVarP(&config.WorkerQueueSize, "worker-size", "s", defaultWorkerQueueSize, "size of each worker queue")
	cmd.Flags().IntVarP(&config.WaitQueueSize, "wait-size", "w", defaultWaitQueueSize, "size of the wait queue")
	cmd.Flags().BoolVarP(&config.Debug, "debug", "d", defaultDebug, "enable debug logs")

	if err := cmd.Execute(); err != nil {
		zap.L().Fatal("execution error", zap.Error(err))
	}
}

func run(cmd *cobra.Command, args []string) error {
	if config.Debug {
		logger, err := zap.NewDevelopment()
		if err != nil {
			log.Panicf("error creating debug logger: %v", err)
		}
		zap.ReplaceGlobals(logger)
	}

	pool := concurrency.NewPool(concurrency.PoolConfig{
		MaxWorkers:      config.MaxWorkers,
		WorkerQueueSize: config.WorkerQueueSize,
		WaitQueueSize:   config.WaitQueueSize,
	})
	pool.Dispatch()

	// TODO: use CSV reader and queue real queries
	queryReader := reader.NewMockQueryReader(20*time.Millisecond, 100*time.Millisecond)
	fakeReadAndQueue(200, 50, queryReader, pool, true)

	result := pool.Wait()
	printPoolResult(result)
	return nil
}

// TODO: pretty print results
func printPoolResult(result concurrency.PoolResult) {
	fmt.Println("\n------------------ Benchmarks -----------------")
	fmt.Printf("total tasks completed: %d\n", result.TotalCompleted)
	fmt.Printf("aggregated duration: %s\n", result.AggregatedDuration.String())
	fmt.Printf("min task duration: %s\n", result.MinTaskDuration.String())
	fmt.Printf("max task duration: %s\n", result.MaxTaskDuration.String())
	fmt.Println("-------------------------------------------------")
}

// TODO: read from csv
func fakeReadAndQueue(numRows int, numHosts int, reader reader.QueryReader, pool *concurrency.Pool, randomAlloc bool) {
	rand.Seed(time.Now().UnixNano())
	hosts := fakeHosts(numHosts)
	for i := 0; i < numRows; i++ {
		var host string

		if randomAlloc {
			host = hosts[rand.Intn(len(hosts))]
		} else {
			// Even distribution across workers
			host = hosts[i%len(hosts)]
		}

		query, _ := reader.ReadRow()
		task := &concurrency.Task{
			RouteKey: host,
			Func:     query,
		}
		pool.Submit(task)
	}
}

// TODO: get hosts from query row
func fakeHosts(numHosts int) []string {
	var hosts []string
	for i := 0; i < numHosts; i++ {
		hosts = append(hosts, fmt.Sprintf("host%d", i))
	}
	return hosts
}
