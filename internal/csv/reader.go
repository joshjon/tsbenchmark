package csv

import (
	"encoding/csv"
	"go.uber.org/zap"
	"io"
)

// Read reads rows from the provided CSV file and sends results to a channel for consumption.
func Read(file io.Reader, bufferSize int) (chan []string, chan error) {
	rowCh := make(chan []string, bufferSize)
	errCh := make(chan error)

	go func() {
		reader := csv.NewReader(file)

		// Read header row
		if _, err := reader.Read(); err != nil {
			errCh <- err
		}

		for {
			row, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					close(rowCh)
					zap.L().Debug("finished reading csv file")
					return
				}
				errCh <- err
				close(errCh)

			}
			rowCh <- row
		}
	}()

	return rowCh, errCh
}
