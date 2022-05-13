package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
		config  Config
	}{
		{
			name:    "negative max workers",
			wantErr: "MaxWorkers: must be no less than 1.",
			config: Config{
				MaxWorkers:         -1,
				WorkerQueueSize:    1,
				WaitQueueSize:      1,
				ReaderBufferSize:   1,
				DatabaseConnection: "non-empty",
			},
		},
		{
			name:    "negative worker queue size",
			wantErr: "WorkerQueueSize: must be no less than 1.",
			config: Config{
				MaxWorkers:         1,
				WorkerQueueSize:    -1,
				WaitQueueSize:      1,
				ReaderBufferSize:   1,
				DatabaseConnection: "non-empty",
			},
		},
		{
			name:    "negative wait queue size",
			wantErr: "WaitQueueSize: must be no less than 1.",
			config: Config{
				MaxWorkers:         1,
				WorkerQueueSize:    1,
				WaitQueueSize:      -1,
				ReaderBufferSize:   1,
				DatabaseConnection: "non-empty",
			},
		},
		{
			name:    "negative reader buffer size",
			wantErr: "ReaderBufferSize: must be no less than 1.",
			config: Config{
				MaxWorkers:         1,
				WorkerQueueSize:    1,
				WaitQueueSize:      1,
				ReaderBufferSize:   -1,
				DatabaseConnection: "non-empty",
			},
		},
		{
			name:    "db connection empty",
			wantErr: "DatabaseConnection: cannot be blank.",
			config: Config{
				MaxWorkers:         1,
				WorkerQueueSize:    1,
				WaitQueueSize:      1,
				ReaderBufferSize:   1,
				DatabaseConnection: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}
