package main

import (
	"database/sql"
	"fmt"
	"github.com/joshjon/tsbenchmark/internal/concurrency"
	"github.com/joshjon/tsbenchmark/internal/csv"
	"github.com/joshjon/tsbenchmark/internal/db"
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
	defaultDBConn           = "host=localhost port=5432 user=postgres password=postgres database=homework"
	defaultDebug            = false
)

var config Config

type Config struct {
	MaxWorkers      int
	WorkerQueueSize int
	WaitQueueSize   int
	Debug           bool
	DBConn          string
}

// TODO:
//  Handle timeouts
//  Use read only and write only channels where applicable
//  add/remove debug logs
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
	cmd.Flags().StringVarP(&config.DBConn, "dbconn", "c", defaultDBConn, "host=x user=x password=x port=x database=x")
	cmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	if config.Debug {
		logger, err := zap.NewDevelopment()
		if err != nil {
			return fmt.Errorf("error creating debug logger: %w", err)
		}
		zap.ReplaceGlobals(logger)
	}

	runStart := time.Now()

	pool := concurrency.NewPool(concurrency.PoolConfig{
		MaxWorkers:      config.MaxWorkers,
		WorkerQueueSize: config.WorkerQueueSize,
		WaitQueueSize:   config.WaitQueueSize,
	})
	pool.Dispatch()

	database, err := db.Open(config.DBConn)
	if err != nil {
		return fmt.Errorf("error opening database connection: %w", err)
	}

	filepath := args[0]
	if err = readAndQueue(filepath, database, pool); err != nil {
		return fmt.Errorf("error reading and queing queries %w", err)
	}

	results := pool.Wait()

	if err = newBenchmark(time.Now().Sub(runStart), results).render(); err != nil {
		return fmt.Errorf("error rendering benchmark results %w", err)
	}

	return nil
}

func readAndQueue(filepath string, database *sql.DB, pool *concurrency.Pool) error {
	csvfile, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening csv file: %w", err)
	}
	defer csvfile.Close()

	rowCh, errCh := csv.Read(csvfile, defaultReaderBufferSize)

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
					_, queryErr := db.QueryMinMaxUsage(database, host, start, end)
					return queryErr
				},
			}
			pool.Submit(task)
		case err = <-errCh:
			return fmt.Errorf("error reading from csv file %w", err)
		}
	}
}

type benchmark struct {
	workersRequired  int
	runtime          time.Duration
	completedQueries int
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
			{Text: pterm.Green("Completed queries: ") + strconv.Itoa(b.completedQueries)},
			{Text: pterm.Green("Total query time: ") + b.totalQueryTime.String()},
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
	}

	b.avgQueryTime = b.totalQueryTime / time.Duration(len(durations))

	if len(durations) > 0 {
		sort.Slice(durations, func(i, j int) bool {
			return durations[i] < durations[j]
		})

		b.medianQueryTime = median(durations)
		b.minQueryTime = durations[0]
		b.maxQueryTime = durations[len(durations)-1]
		b.avgQueryTime = b.totalQueryTime / time.Duration(len(durations))
	}

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
