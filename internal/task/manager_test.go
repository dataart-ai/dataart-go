package task

import (
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	_, err := NewManager(0, 1, 1, 1, nil, nil)
	if err == nil {
		t.Error("given numWorkers is invalid")
		t.Fail()
	}

	_, err = NewManager(1, 0, 1, 1, nil, nil)
	if err == nil {
		t.Error("given bufferSize is invalid")
		t.Fail()
	}

	_, err = NewManager(1, 1, -1, 1, nil, nil)
	if err == nil {
		t.Error("given numRetries is invalid")
		t.Fail()
	}

	_, err = NewManager(1, 1, 1, 0, nil, nil)
	if err == nil {
		t.Error("given backoffRatio is invalid")
		t.Fail()
	}
}

func TestManager_WithWorkersLessThanTasks(t *testing.T) {
	t.Parallel()

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
		tm.Queue(func() error {
			time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
			return nil
		})
	}

	tm.Shutdown()
	if doneTasks != numTasks {
		t.Fail()
	}
}

func TestManager_WithWorkersEqualToTasks(t *testing.T) {
	t.Parallel()

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
		tm.Queue(func() error {
			time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
			return nil
		})
	}

	tm.Shutdown()
	if doneTasks != numTasks {
		t.Fail()
	}
}

func TestManager_WithWorkersMoreThanTasks(t *testing.T) {
	t.Parallel()

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
		tm.Queue(func() error {
			time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
			return nil
		})
	}

	tm.Shutdown()
	if doneTasks != numTasks {
		t.Fail()
	}
}

func TestManager_WithFailingTaskShouldRetryEnoughTimes(t *testing.T) {
	t.Parallel()

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
		tm.Queue(func() error {
			time.Sleep(time.Duration(rand.Int31n(10)) * time.Millisecond)
			return errors.New("tasks failed for some reason")
		})
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
	t.Parallel()

	tm, _ := NewManager(1, 1, 1, 1, nil, nil)

	tm.Queue(func() error {
		return nil
	})

	// Multiple shutdowns should be idempotent.
	tm.Shutdown()
	tm.Shutdown()

	err := tm.Queue(func() error {
		return nil
	})

	if err == nil {
		t.Error("queue after shutdown should have returned an error")
		t.Fail()
	}
}
