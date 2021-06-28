package task

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dataart-ai/dataart-go/internal/pkg/atomicutil"
)

// Manager receives task functions and distributes them among worker goroutines.
// If any given function returns an error it will be retried when numRetries > 0.
type Manager struct {
	numWorkers   int
	bufferSize   int
	numRetries   int
	backoffRatio int

	doneHook func(taskUID string, workerID string)
	failHook func(taskUID string, workerID string, err error)

	buffer chan task

	once       sync.Once
	wg         sync.WaitGroup
	inShutdown atomicutil.Bool
	isStarted  atomicutil.Bool
}

func (m *Manager) start() {
	m.isStarted.SetTrue()

	for i := 0; i < m.numWorkers; i++ {
		go func(wid, numRetries, backoffRatio int, wg *sync.WaitGroup, buffer <-chan task,
			doneHook func(taskUID string, workerID string),
			failHook func(taskUID string, workerID string, err error)) {

			workerID := fmt.Sprintf("worker-%d", wid)

			for t := range buffer {
				// We add 1 to numRetries for the first run.
				for r := 0; r < numRetries+1; r++ {
					err := t.work()
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
		}(i, m.numRetries, m.backoffRatio, &m.wg, m.buffer, m.doneHook, m.failHook)
	}
}

// Queue enqueues given function to be executed. If given returns an error
// it will be retried numRetries times until giving up. Queue returns an error
// if the manager instance is shutting down.
func (m *Manager) Queue(work func() error) error {
	if m.inShutdown.IsSet() {
		return errors.New("manager is shutting down")
	}

	m.once.Do(m.start)

	m.wg.Add(1)
	m.buffer <- newTask(work)
	return nil
}

// Shutdown terminates Manager gracefully. It waits for all workers to return
// then closes the buffer channel and returns.
func (m *Manager) Shutdown() {
	if m.inShutdown.IsSet() || !m.isStarted.IsSet() {
		return
	}
	m.inShutdown.SetTrue()

	m.wg.Wait()
	close(m.buffer)
}

// NewManager creates a new Manager instance using provided values. Use this
// function to instantiate a concrete Manager type.
func NewManager(numWorkers, bufferSize, numRetries, backoffRatio int,
	doneHook func(taskUID, workerID string),
	failHook func(taskUID, workerID string, err error)) (*Manager, error) {

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

	tm := &Manager{
		numWorkers:   numWorkers,
		bufferSize:   bufferSize,
		numRetries:   numRetries,
		backoffRatio: backoffRatio,
		doneHook:     doneHook,
		failHook:     failHook,
		buffer:       make(chan task, bufferSize),
	}

	return tm, nil
}
