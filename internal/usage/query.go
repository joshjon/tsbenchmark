package usage

import (
	"database/sql"
)

const query = `SELECT time_bucket('1 minutes', ts) AS bucket, MIN(usage) AS min_usage, MAX(usage) AS max_usage, host, COUNT(*)
FROM cpu_usage
WHERE host = $1 and ts >= $2 and ts <= $3
GROUP BY bucket, host;`

type Result struct {
	Count    int
	Interval string
	Host     string
	Min      float64
	Max      float64
}

// QueryMinMaxUsagePerMinuteInRange returns the max cpu usage and min cpu usage of the given
// hostname for every minute in the time range specified by the start time and end time.
func QueryMinMaxUsagePerMinuteInRange(db *sql.DB, host string, startTimestamp string, endTimestamp string) ([]Result, error) {
	rows, err := db.Query(query, host, startTimestamp, endTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Result

	for rows.Next() {
		var result Result

		if err = rows.Scan(&result.Interval, &result.Min, &result.Max, &result.Host, &result.Count); err != nil {
			return nil, err
		}

		items = append(items, result)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
