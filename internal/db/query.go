package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type UsageResult struct {
	Min float64
	Max float64
}

func QueryMinMaxUsage(db *sql.DB, host string, startTimestamp string, endTimestamp string) (*UsageResult, error) {
	query := fmt.Sprintf(
		"select MIN(usage), MAX(usage) "+
			"from cpu_usage "+
			"where host = '%s' and ts >= '%s' and ts <= '%s'",
		host, startTimestamp, endTimestamp,
	)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var result UsageResult

		err = rows.Scan(&result.Min, &result.Max)
		if err != nil {
			return nil, err
		}

		return &result, nil
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return nil, errors.New("no row returned from query")
}
