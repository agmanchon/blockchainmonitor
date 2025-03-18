package usecases

import (
	"context"
	"log"
	"sync"
)

type Worker interface {
	// Start worker background process
	Start(ctx context.Context, waitGroup *sync.WaitGroup)
}

// WorkerManager define worker manager functionality
type WorkerManager interface {
	// Add adds new worker to the worker manager
	Add(worker Worker)
	// AddFunc adds new function that can be invoked as a worker
	AddFunc(workerFunc StartFunc)
	// Start starts all workers in the worker manager
	Start(ctx context.Context, waitGroup *sync.WaitGroup)
}

// DefaultWorkerManager default worker manager implementation that handle a collection of workers
type DefaultWorkerManager struct {
	workerFuncs []StartFunc
}

// StartFunc is the func interface for a process that can be handled as a worker
type StartFunc func(ctx context.Context, waitGroup *sync.WaitGroup)

// NewDefaultWorkerManager creates aDefaultWorkerManager
func NewDefaultWorkerManager() *DefaultWorkerManager {
	return &DefaultWorkerManager{
		workerFuncs: make([]StartFunc, 0),
	}
}

// Add adds new worker to the worker manager
func (manager *DefaultWorkerManager) Add(worker Worker) {
	manager.AddFunc(worker.Start)

}

// AddFunc adds new function that can be invoked as a worker
func (manager *DefaultWorkerManager) AddFunc(workerFunc StartFunc) {
	manager.workerFuncs = append(manager.workerFuncs, workerFunc)
}

// Start starts all workers in the worker manager
func (manager *DefaultWorkerManager) Start(ctx context.Context, mainWaitGroup *sync.WaitGroup) {
	defer mainWaitGroup.Done()
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(len(manager.workerFuncs))

	for i := range manager.workerFuncs {
		go manager.workerFuncs[i](ctx, waitGroup)
	}

	waitGroup.Wait()
	log.Println("All workers finished")
}
