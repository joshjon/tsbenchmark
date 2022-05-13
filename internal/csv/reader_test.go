package csv

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestRead(t *testing.T) {
	csvfile, err := os.Open("testdata/valid.csv")
	require.NoError(t, err)
	defer csvfile.Close()

	rowsCh, errCh := Read(csvfile, 1)

	wantRows := 5
	wantColumns := 3

	for i := 0; i < wantRows; i++ {
		row := <-rowsCh
		assert.Len(t, row, wantColumns)
	}

	_, open := <-rowsCh
	assert.False(t, open)
	assert.Empty(t, errCh)
}

func TestRead_error(t *testing.T) {
	wantErr := errors.New("some error")

	file := errFile{
		wantErr: wantErr,
	}

	rowsCh, errCh := Read(file, 1)
	err := <-errCh
	assert.EqualError(t, err, wantErr.Error())
	assert.Empty(t, rowsCh)
}

type errFile struct {
	wantErr error
}

func (e errFile) Read(_ []byte) (int, error) {
	return 0, e.wantErr
}
