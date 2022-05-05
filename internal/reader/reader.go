package reader

import (
	"math/rand"
	"time"
)

type Query func() time.Duration

type QueryReader interface {
	ReadRow() (query Query, more bool)
}

type CSVQueryReader struct {
	file string
}

func NewCSVQueryReader(file string) *CSVQueryReader {
	// TODO
	return &CSVQueryReader{
		file: file,
	}
}

func (r *CSVQueryReader) ReadRow() Query {
	// TODO
	return nil
}

// MockQueryReader TODO: move to test file
type MockQueryReader struct {
	min time.Duration
	max time.Duration
}

func NewMockQueryReader(min time.Duration, max time.Duration) *MockQueryReader {
	return &MockQueryReader{
		min: min,
		max: max,
	}
}

func (r *MockQueryReader) ReadRow() (Query, bool) {
	return func() time.Duration {
		rand.Seed(time.Now().UnixNano())
		duration := rand.Intn(int(r.max-r.min)) + int(r.min)
		start := time.Now()
		time.Sleep(time.Duration(duration))
		return time.Now().Sub(start)
	}, true
}
