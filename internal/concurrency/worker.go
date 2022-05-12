package concurrency

import (
	"github.com/fatih/set"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Task struct {
	RouteKey string
	Func     func() (err error)
}

type WorkerResult struct {
	sync.Mutex
	ID            int
	Completed     int
	TotalDuration time.Duration
	TaskDurations []time.Duration
}

type WorkerConfig struct {
	QueueSize int
}

type Worker struct {
	ID           int
	done         chan *WorkerResult
	workerQueue  chan *Task
	taskQueue    <-chan *Task
	workerResult *WorkerResult
	routeKeys    set.Interface
}

func NewWorker(queueSize int, taskQueue <-chan *Task) *Worker {
	workerID := time.Now().Nanosecond()
	return &Worker{
		routeKeys:   set.New(set.ThreadSafe),
		ID:          workerID,
		done:        make(chan *WorkerResult),
		taskQueue:   taskQueue,
		workerQueue: make(chan *Task, queueSize),
		workerResult: &WorkerResult{
			ID: workerID,
		},
	}
}

// Start starts the worker in a new goroutine and receives tasks from the worker queue.
// If the worker queue is empty then tasks will be received from the task queue.
func (w *Worker) Start() {
	go func() {
		for {
			// Due to the random nature of select statements, a single worker queue case
			//is required to ensure the worker queue is prioritised over the task queue.
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
					// Task queue is closed in which it is safe to stop the worker.
					w.done <- w.workerResult
					close(w.done)
					return
				}
				w.routeKeys.Add(task.RouteKey)
				w.execute(task)
			}
		}
	}()
}

func (w *Worker) Enqueue(task *Task) {
	w.workerQueue <- task
}

func (w *Worker) HasRouteKey(routeKey string) bool {
	return w.routeKeys.Has(routeKey)
}

func (w *Worker) Wait() *WorkerResult {
	return <-w.done
}

func (w *Worker) execute(task *Task) {
	start := time.Now()

	// TODO: handle gracefully by sending to an err chan
	err := task.Func()
	if err != nil {
		zap.L().Panic("task error", zap.Error(err))
	}

	duration := time.Now().Sub(start)
	w.workerResult.Completed += 1
	w.workerResult.TotalDuration += duration
	w.workerResult.TaskDurations = append(w.workerResult.TaskDurations, duration)
}
