package main

import (
	"database/sql"
	"fmt"
	"github.com/joshjon/tsbenchmark/internal/concurrency"
	"github.com/joshjon/tsbenchmark/internal/config"
	"github.com/joshjon/tsbenchmark/internal/csv"
	"github.com/joshjon/tsbenchmark/internal/db"
	"github.com/joshjon/tsbenchmark/internal/usage"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"sort"
	"strconv"
	"time"
)

const (
	defaultMaxWorkers       = 10
	defaultWorkerQueueSize  = 50
	defaultWaitQueueSize    = 500
	defaultReaderBufferSize = 500
	defaultDBConn           = "host=timescaledb port=5432 user=postgres password=postgres database=homework"
	defaultDebug            = false
)

var cfg config.Config

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

	cmd.Flags().IntVarP(&cfg.MaxWorkers, "max-workers", "m", defaultMaxWorkers, "max number of concurrent workers")
	cmd.Flags().IntVarP(&cfg.WorkerQueueSize, "worker-size", "s", defaultWorkerQueueSize, "size of each worker queue")
	cmd.Flags().IntVarP(&cfg.WaitQueueSize, "wait-size", "w", defaultWaitQueueSize, "size of the wait queue")
	cmd.Flags().IntVarP(&cfg.ReaderBufferSize, "reader-size", "r", defaultReaderBufferSize, "size of the file reader buffer")
	cmd.Flags().BoolVarP(&cfg.Debug, "debug", "d", defaultDebug, "enable debug logs")
	cmd.Flags().StringVarP(&cfg.DatabaseConnection, "dbconn", "c", defaultDBConn, "host=x user=x password=x port=x database=x")
	cmd.Execute()
}

// run creates a new worker pool and starts dispatching any received tasks to its workers in the background.
// Rows are read from the specified CPU usage CSV file and transformed into queries that return the max and
// min cpu of the host for every minute between the start end time. Each query is submitted as a task to the
// worker pool task queue which are then picked up and executed by workers. A route key is used to ensure all
// queries with a particular host name are executed on the same worker. Finally, wait occurs until all query
// tasks have been completed.
func run(cmd *cobra.Command, args []string) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	if cfg.Debug {
		logger, err := zap.NewDevelopment()
		if err != nil {
			return fmt.Errorf("error creating debug logger: %w", err)
		}
		zap.ReplaceGlobals(logger)
	}

	runStart := time.Now()

	pool := concurrency.NewPool(concurrency.PoolConfig{
		MaxWorkers:      cfg.MaxWorkers,
		WorkerQueueSize: cfg.WorkerQueueSize,
		WaitQueueSize:   cfg.WaitQueueSize,
	})
	pool.Dispatch()

	database, err := db.Open(cfg.DatabaseConnection)
	if err != nil {
		return fmt.Errorf("error opening database connection: %w", err)
	}

	filepath := args[0]
	if err = readAndQueue(filepath, database, pool); err != nil {
		return fmt.Errorf("error reading and queing queries: %w", err)
	}

	results := pool.Wait()

	if err = newBenchmark(time.Now().Sub(runStart), results).render(); err != nil {
		return fmt.Errorf("error rendering benchmark results: %w", err)
	}

	return nil
}

func readAndQueue(filepath string, database *sql.DB, pool *concurrency.Pool) error {
	csvfile, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening csv file: %w", err)
	}
	defer csvfile.Close()

	rowCh, errCh := csv.Read(csvfile, cfg.ReaderBufferSize)

	for {
		select {
		case row, ok := <-rowCh:
			if !ok {
				return nil
			}
			host, start, end := row[0], row[1], row[2]

			task := &concurrency.Task{
				RouteKey: host,
				Func: func() error {
					_, queryErr := usage.QueryMinMaxUsagePerMinuteInRange(database, host, start, end)
					return queryErr
				},
			}
			pool.Submit(task)
		case err = <-errCh:
			return fmt.Errorf("error reading from csv file: %w", err)
		}
	}
}

type benchmark struct {
	workersRequired  int
	runtime          time.Duration
	completedQueries int
	erroredQueries   int
	totalQueryTime   time.Duration
	minQueryTime     time.Duration
	maxQueryTime     time.Duration
	medianQueryTime  time.Duration
	avgQueryTime     time.Duration
}

func (b benchmark) render() error {
	header := pterm.NewStyle(pterm.FgWhite, pterm.BgDarkGray, pterm.Bold)
	header.Println("\n                   Benchmarks                   ")

	return pterm.DefaultBulletList.WithItems(
		[]pterm.BulletListItem{
			{Text: pterm.Green("Workers started: ") + strconv.Itoa(b.workersRequired)},
			{Text: pterm.Green("Runtime: ") + b.runtime.String()},
			{Text: pterm.Green("Query processing time (across workers): ") + b.totalQueryTime.String()},
			{Text: pterm.Green("Query executions: ") + strconv.Itoa(b.completedQueries)},
			{Text: pterm.Green("Query errors: ") + strconv.Itoa(b.erroredQueries)},
			{Text: pterm.Green("Min query time: ") + b.minQueryTime.String()},
			{Text: pterm.Green("Max query time: ") + b.maxQueryTime.String()},
			{Text: pterm.Green("Median query time: ") + b.medianQueryTime.String()},
			{Text: pterm.Green("Average query time: ") + b.avgQueryTime.String()},
		},
	).Render()
}

func newBenchmark(runtime time.Duration, results []*concurrency.WorkerResult) benchmark {
	b := benchmark{
		workersRequired: len(results),
		runtime:         runtime,
	}

	var durations []time.Duration
	for _, result := range results {
		b.completedQueries += result.Completed
		durations = append(durations, result.TaskDurations...)
		b.totalQueryTime += result.TotalDuration
		b.erroredQueries += len(result.Errors)

		for _, taskErr := range result.Errors {
			zap.L().Error("query error", zap.Error(taskErr))
		}
	}

	if len(durations) == 0 {
		return benchmark{}
	}

	if len(durations) == 1 {
		b.avgQueryTime = durations[0]
		b.medianQueryTime = durations[0]
	}

	if len(durations) > 1 {
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		b.avgQueryTime = b.totalQueryTime / time.Duration(len(durations))
		b.medianQueryTime = median(durations)
	}

	b.minQueryTime = durations[0]
	b.maxQueryTime = durations[len(durations)-1]

	return b
}

func median(nums []time.Duration) time.Duration {
	i := len(nums) / 2
	m := nums[i]
	if i%2 == 0 {
		m = (m + nums[i+1]) / 2
	}
	return m
}
