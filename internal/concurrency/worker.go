package concurrency

import (
	"github.com/fatih/set"
	"go.uber.org/zap"
	"time"
)

type Task struct {
	RouteKey string
	Func     func() error
}

type WorkerResult struct {
	Completed     int
	TotalDuration time.Duration
	TaskDurations []time.Duration
	Errors        []error
}

type WorkerConfig struct {
	QueueSize int
}

type Worker struct {
	done         chan *WorkerResult
	workerQueue  chan *Task
	taskQueue    <-chan *Task
	workerResult *WorkerResult
	routeKeys    set.Interface
}

func NewWorker(queueSize int, taskQueue <-chan *Task) *Worker {
	return &Worker{
		routeKeys:    set.New(set.ThreadSafe),
		done:         make(chan *WorkerResult),
		taskQueue:    taskQueue,
		workerQueue:  make(chan *Task, queueSize),
		workerResult: &WorkerResult{},
	}
}

// Start continuously receives tasks from the worker queue to execute as first priority.
// If the worker queue is empty, tasks will be pulled from the task queue instead.
// Finally, when the worker queue is empty and the task queue is closed, the worker result
// is sent to the done channel to indicate completion.
func (w *Worker) Start() {
	go func() {
		for {
			// Due to the random nature of select statements, a single case is required
			// to ensure the worker queue is prioritised over the task queue.
			select {
			case task := <-w.workerQueue:
				w.execute(task)
				continue
			default:
			}

			select {
			case task := <-w.workerQueue:
				w.execute(task)
				continue
			case task, ok := <-w.taskQueue:
				if !ok {
					w.done <- w.workerResult
					close(w.workerQueue)
					close(w.done)
					zap.L().Debug("task queue closed, worker done")
					return
				}
				w.routeKeys.Add(task.RouteKey)
				w.execute(task)
			}
		}
	}()
}

// Submit adds a task to the worker queue.
func (w *Worker) Submit(task *Task) {
	w.workerQueue <- task
}

// HasRouteKey checks if the worker has been allocated to the given route key.
func (w *Worker) HasRouteKey(routeKey string) bool {
	return w.routeKeys.Has(routeKey)
}

// Wait blocks for the worker result to be received.
func (w *Worker) Wait() *WorkerResult {
	return <-w.done
}

func (w *Worker) execute(task *Task) {
	start := time.Now()

	if err := task.Func(); err != nil {
		w.workerResult.Errors = append(w.workerResult.Errors, err)
	}

	duration := time.Now().Sub(start)
	w.workerResult.Completed += 1
	w.workerResult.TotalDuration += duration
	w.workerResult.TaskDurations = append(w.workerResult.TaskDurations, duration)
}
