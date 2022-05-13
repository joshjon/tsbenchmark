package concurrency

import (
	"go.uber.org/zap"
	"sync"
)

type PoolConfig struct {
	MaxWorkers      int
	WorkerQueueSize int
	WaitQueueSize   int
}

type Pool struct {
	config    PoolConfig
	waitQueue chan *Task
	taskQueue chan *Task
	workers   poolWorkers
	done      chan bool
}

func NewPool(config PoolConfig) *Pool {
	return &Pool{
		config:    config,
		waitQueue: make(chan *Task, config.WaitQueueSize),
		taskQueue: make(chan *Task),
		done:      make(chan bool),
	}
}

// Dispatch expects tasks to be submitted to the wait queue with Submit. Tasks are received
// from the queue and are then directed as follows. If a worker exists and the task route key
// is allocated to it, add the task to the worker's queue. Otherwise, start a new worker (if
// not at max) and add the task to the task queue so that any available worker can receive it.
// It is worth noting that workers are only started when an available task with an unallocated
// route key is received, which ensures that workers are not unnecessarily spun up. For example,
// a task queue of 100 tasks with identical route keys must be routed to the same worker, so we
// only start 1 worker even if the max allows for more.
func (p *Pool) Dispatch() {
	go func() {
		for task := range p.waitQueue {
			if worker, ok := p.workers.findByRouteKey(task.RouteKey); ok {
				worker.Submit(task)
			} else {
				if p.workers.len() < p.config.MaxWorkers {
					w := NewWorker(p.config.WorkerQueueSize, p.taskQueue)
					w.Start()
					p.workers.append(w)
					zap.L().Debug("worker started", zap.Int("worker_count", p.workers.len()))
				}
				p.taskQueue <- task
			}
		}

		close(p.taskQueue)
		p.done <- true
		zap.L().Debug("wait queue closed, dispatch done")
	}()
}

// Submit adds a task to the pool's wait queue.
func (p *Pool) Submit(task *Task) {
	p.waitQueue <- task
}

// Wait blocks until all tasks have completed and until all worker results have been received.
func (p *Pool) Wait() []*WorkerResult {
	close(p.waitQueue)
	<-p.done

	for {
		if len(p.waitQueue) == 0 && len(p.taskQueue) == 0 {
			break
		}
	}

	return p.workers.waitAll()
}

type poolWorkers struct {
	sync.Mutex
	workers []*Worker
}

func (l *poolWorkers) append(worker *Worker) {
	l.Lock()
	defer l.Unlock()
	l.workers = append(l.workers, worker)
}

func (l *poolWorkers) findByRouteKey(routeKey string) (*Worker, bool) {
	l.Lock()
	defer l.Unlock()
	for _, worker := range l.workers {
		if worker.HasRouteKey(routeKey) {
			return worker, true
		}
	}
	return nil, false
}

func (l *poolWorkers) waitAll() []*WorkerResult {
	l.Lock()
	defer l.Unlock()
	var results []*WorkerResult
	for _, worker := range l.workers {
		results = append(results, worker.Wait())
	}
	return results
}

func (l *poolWorkers) len() int {
	l.Lock()
	defer l.Unlock()
	return len(l.workers)
}
