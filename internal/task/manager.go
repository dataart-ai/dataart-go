package task

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dataart-ai/dataart-go/internal/pkg/atomicutil"
)

type Manager interface {
	Queue(t Task) error
	Shutdown()
}

type managerImpl struct {
	numWorkers   int
	bufferSize   int
	numRetries   int
	backoffRatio int

	doneHook func(taskUID string, workerID string)
	failHook func(taskUID string, workerID string, err error)

	buffer chan Task

	once       sync.Once
	wg         sync.WaitGroup
	inShutdown atomicutil.Bool
	isStarted  atomicutil.Bool
}

func (tm *managerImpl) start() {
	tm.isStarted.SetTrue()

	for i := 0; i < tm.numWorkers; i++ {
		go func(wid, numRetries, backoffRatio int, wg *sync.WaitGroup, buffer <-chan Task,
			doneHook func(taskUID string, workerID string),
			failHook func(taskUID string, workerID string, err error)) {

			workerID := fmt.Sprintf("worker-%d", wid)

			for t := range buffer {
				// We add 1 to numRetries for the first run.
				for r := 0; r < numRetries+1; r++ {
					err := t.Work()
					if err != nil {
						// Job failed. Worker will sleep for (backoffRatio*r) seconds and retry.
						if failHook != nil {
							failHook(t.id, workerID, err)
						}

						// We add 1 to r since it starts with 0.
						backoff := backoffRatio * (r + 1)
						time.Sleep(time.Duration(backoff) * time.Second)
					} else if err == nil {
						if doneHook != nil {
							doneHook(t.id, workerID)
						}

						break
					}
				}

				wg.Done()
			}
		}(i, tm.numRetries, tm.backoffRatio, &tm.wg, tm.buffer, tm.doneHook, tm.failHook)
	}
}

func (tm *managerImpl) Queue(t Task) error {
	if tm.inShutdown.IsSet() {
		return errors.New("manager is shutting down")
	}

	tm.once.Do(tm.start)

	tm.wg.Add(1)
	tm.buffer <- t
	return nil
}

func (tm *managerImpl) Shutdown() {
	if tm.inShutdown.IsSet() || !tm.isStarted.IsSet() {
		return
	}
	tm.inShutdown.SetTrue()

	tm.wg.Wait()
	close(tm.buffer)
}

func NewManager(numWorkers, bufferSize, numRetries, backoffRatio int,
	doneHook func(taskUID, workerID string),
	failHook func(taskUID, workerID string, err error)) (Manager, error) {

	if numWorkers < 1 {
		return nil, errors.New("numWorkers must be at least 1")
	}

	if bufferSize < 1 {
		return nil, errors.New("bufferSize must be at least 1")
	}

	if numRetries < 0 {
		return nil, errors.New("numRetries can't be negative")
	}

	if backoffRatio < 1 {
		return nil, errors.New("backoffRatio must be at least 1")
	}

	tm := &managerImpl{
		numWorkers:   numWorkers,
		bufferSize:   bufferSize,
		numRetries:   numRetries,
		backoffRatio: backoffRatio,
		doneHook:     doneHook,
		failHook:     failHook,
		buffer:       make(chan Task, bufferSize),
	}

	return tm, nil
}
