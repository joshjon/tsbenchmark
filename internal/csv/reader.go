package csv

import (
	"encoding/csv"
	"io"
)

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
					return
				}
				errCh <- err

			}
			rowCh <- row
		}
	}()

	return rowCh, errCh
}
