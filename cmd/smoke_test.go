//go:build smoke

package main

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	queryParamFile = "../database/test_query_param.csv"
	dbConn         = "host=localhost port=5432 user=postgres password=postgres database=homework"
)

func Test_run(t *testing.T) {
	cfg.MaxWorkers = 1
	cfg.WaitQueueSize = 1
	cfg.WorkerQueueSize = 1
	cfg.ReaderBufferSize = 10
	cfg.DatabaseConnection = dbConn

	cmd := &cobra.Command{
		RunE: run,
	}

	err := run(cmd, []string{queryParamFile})
	require.NoError(t, err)
}
