package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
		config  Config
		fields  []string
	}{
		{
			name:    "field required",
			wantErr: "cannot be blank",
			config: Config{
				MaxWorkers:         0,
				WorkerQueueSize:    0,
				WaitQueueSize:      0,
				ReaderBufferSize:   0,
				DatabaseConnection: "",
			},
			fields: []string{"MaxWorkers", "WorkerQueueSize", "WaitQueueSize", "ReaderBufferSize", "DatabaseConnection"},
		},
		{
			name:    "int must be positive",
			wantErr: "must be no less than 1",
			config: Config{
				MaxWorkers:         -1,
				WorkerQueueSize:    -1,
				WaitQueueSize:      -1,
				ReaderBufferSize:   -1,
				DatabaseConnection: "non-empty",
			},
			fields: []string{"MaxWorkers", "WorkerQueueSize", "WaitQueueSize", "ReaderBufferSize"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			for _, field := range tt.fields {
				assert.Contains(t, err.Error(), fmt.Sprintf("%s: %s", field, tt.wantErr))
			}
		})
	}
}
