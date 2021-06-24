package task

import (
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	// numWorkers value
	_, err := NewManager(0, 1, 1, 1, nil, nil)
	if err == nil {
		t.Fail()
	}

	// bufferSize value
	_, err = NewManager(1, 0, 1, 1, nil, nil)
	if err == nil {
		t.Fail()
	}

	// numRetries value
	_, err = NewManager(1, 1, -1, 1, nil, nil)
	if err == nil {
		t.Fail()
	}

	// backoffRatio value
	_, err = NewManager(1, 1, 1, 0, nil, nil)
	if err == nil {
		t.Fail()
	}
}

func TestManager_WithWorkersLessThanTasks(t *testing.T) {
	numTasks := 5
	doneTasks := 0
	numWorkers := 2
	mx := sync.Mutex{}
	doneHook := func(tid, wid string) {
		mx.Lock()
		doneTasks += 1
		mx.Unlock()
	}
	tm, _ := NewManager(numWorkers, 1, 1, 1, doneHook, nil)

	for i := 0; i < numTasks; i++ {
		t := NewTask(func() error {
			time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
			return nil
		})
		tm.Queue(t)
	}

	tm.Shutdown()
	if doneTasks != numTasks {
		t.Fail()
	}
}

func TestManager_WithWorkersEqualToTasks(t *testing.T) {
	numTasks := 5
	doneTasks := 0
	numWorkers := 5
	mx := sync.Mutex{}
	doneHook := func(tid, wid string) {
		mx.Lock()
		doneTasks += 1
		mx.Unlock()
	}
	tm, _ := NewManager(numWorkers, 1, 1, 1, doneHook, nil)

	for i := 0; i < numTasks; i++ {
		t := NewTask(func() error {
			time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
			return nil
		})
		tm.Queue(t)
	}

	tm.Shutdown()
	if doneTasks != numTasks {
		t.Fail()
	}
}

func TestManager_WithWorkersMoreThanTasks(t *testing.T) {
	numTasks := 2
	doneTasks := 0
	numWorkers := 5
	mx := sync.Mutex{}
	doneHook := func(tid, wid string) {
		mx.Lock()
		doneTasks += 1
		mx.Unlock()
	}
	tm, _ := NewManager(numWorkers, 1, 1, 1, doneHook, nil)

	for i := 0; i < numTasks; i++ {
		t := NewTask(func() error {
			time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
			return nil
		})
		tm.Queue(t)
	}

	tm.Shutdown()
	if doneTasks != numTasks {
		t.Fail()
	}
}

func TestManager_WithFailingTask(t *testing.T) {
	numTasks := 3
	numTries := 0
	numWorkers := 2
	numRetries := 2
	mx := sync.Mutex{}
	failHook := func(tid, wid string, err error) {
		mx.Lock()
		numTries += 1
		mx.Unlock()
	}
	tm, _ := NewManager(numWorkers, 1, numRetries, 1, nil, failHook)

	for i := 0; i < numTasks; i++ {
		t := NewTask(func() error {
			time.Sleep(time.Duration(rand.Int31n(10)) * time.Millisecond)
			return errors.New("tasks failed for some reason")
		})
		tm.Queue(t)
	}

	tm.Shutdown()
	// We have to account for the first each task is executed (numTasks) plus
	//  the times they are being retried (numTasks * numRetries).
	triesShouldBe := numTasks + (numTasks * numRetries)
	if numTries != triesShouldBe {
		t.Errorf("numTries should have been %d, instead got %d", triesShouldBe, numTries)
		t.Fail()
	}
}

func TestManager_WithQueueAfterShutdown(t *testing.T) {
	tm, _ := NewManager(1, 1, 1, 1, nil, nil)

	tm.Queue(NewTask(func() error {
		return nil
	}))

	// Multiple shutdowns should be idempotent.
	tm.Shutdown()
	tm.Shutdown()

	err := tm.Queue(NewTask(func() error {
		return nil
	}))

	if err == nil {
		t.Fail()
	}
}
