package usage

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueryMinMaxUsagePerMinuteInRange(t *testing.T) {
	host := "host_000008"
	start := "2017-01-02 18:50:28"
	end := "2017-01-02 19:50:28"
	wantInterval := "2017-01-02 18:51:28"
	wantMin := float64(20)
	wantMax := float64(40)
	wantCount := 1

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"time", "min", "max", "host", "count"}).
		AddRow(wantInterval, float64(20), float64(40), host, 1)

	mock.ExpectQuery(".*").WillReturnRows(rows)
	results, err := QueryMinMaxUsagePerMinuteInRange(db, host, start, end)
	require.NoError(t, err)
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, host, result.Host)
	assert.Equal(t, wantInterval, result.Interval)
	assert.Equal(t, wantMin, result.Min)
	assert.Equal(t, wantMax, result.Max)
	assert.Equal(t, wantCount, result.Count)
}
