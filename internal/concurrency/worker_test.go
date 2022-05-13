package concurrency

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPool_Worker_taskSuccess(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "task execution success",
			wantErr: nil,
		},
		{
			name:    "task execution error",
			wantErr: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskQueue := make(chan *Task)
			worker := NewWorker(10, taskQueue)
			worker.Start()
			task := &Task{
				Func: func() error {
					return tt.wantErr
				},
			}
			worker.Submit(task)
			close(taskQueue)
			result := worker.Wait()
			assert.Equal(t, 1, result.Completed)
			assert.NotEmpty(t, result.TaskDurations)
			assert.NotEmpty(t, result.TotalDuration)

			if tt.wantErr != nil {
				assert.EqualError(t, result.Errors[0], tt.wantErr.Error())
			} else {
				assert.Empty(t, result.Errors)
			}
		})
	}
}
