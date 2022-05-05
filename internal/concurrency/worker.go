package concurrency

import (
	"github.com/fatih/set"
	"time"
)

type Task struct {
	RouteKey string
	Func     func() time.Duration
}

type WorkerConfig struct {
	QueueSize int
}

type Worker struct {
	ID               int
	workerQueue      chan *Task
	taskQueue        <-chan *Task
	completionFuture *CompletionFuture
	routeKeys        set.Interface
}

func NewWorker(queueSize int, taskQueue <-chan *Task) *Worker {
	workerID := time.Now().Nanosecond()
	return &Worker{
		routeKeys:        set.New(set.ThreadSafe),
		ID:               workerID,
		taskQueue:        taskQueue,
		workerQueue:      make(chan *Task, queueSize),
		completionFuture: newCompletionFuture(workerID),
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
				w.completionFuture.addTaskDuration(task.Func())
				continue
			default:
			}

			select {
			case task := <-w.workerQueue:
				w.completionFuture.addTaskDuration(task.Func())
				continue
			case task, ok := <-w.taskQueue:
				if !ok {
					// Task queue is closed in which it is safe to stop the worker.
					w.completionFuture.send()
					return
				}
				w.routeKeys.Add(task.RouteKey)
				w.completionFuture.addTaskDuration(task.Func())
			}
		}
	}()
}

func (w *Worker) Enqueue(task *Task) {
	w.workerQueue <- task
}

func (w *Worker) CompletionFuture() *CompletionFuture {
	return w.completionFuture
}

func (w *Worker) HasRouteKey(routeKey string) bool {
	return w.routeKeys.Has(routeKey)
}
