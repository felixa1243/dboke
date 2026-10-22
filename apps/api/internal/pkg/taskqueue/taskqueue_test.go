package taskqueue

import (
	"errors"
	"testing"
	"time"
)

// resetForTest installs a fresh queue with the given capacity and no workers.
func resetForTest(capacity int) {
	mu.Lock()
	tasks = make(map[string]*Task)
	mu.Unlock()
	jobQueue = make(chan func(), capacity)
}

func TestEnqueueTask_FullQueueDoesNotBlock(t *testing.T) {
	resetForTest(2)

	for i := 0; i < 2; i++ {
		if _, err := EnqueueTask("queued", func(string) {}); err != nil {
			t.Fatalf("enqueue %d: %v", i, err)
		}
	}

	done := make(chan error, 1)
	go func() {
		_, err := EnqueueTask("overflow", func(string) {})
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, ErrQueueFull) {
			t.Fatalf("err = %v, want ErrQueueFull", err)
		}
	case <-time.After(time.Second):
		t.Fatal("EnqueueTask blocked on a full queue")
	}

	mu.RLock()
	defer mu.RUnlock()
	if len(tasks) != 2 {
		t.Errorf("%d tasks tracked, want 2 (rejected task must not be recorded)", len(tasks))
	}
}

func TestFinishedTasksArePruned(t *testing.T) {
	resetForTest(10)

	oldID, _ := EnqueueTask("old", func(string) {})
	runningID, _ := EnqueueTask("running", func(string) {})
	UpdateTask(oldID, StatusCompleted, 100, "", "")
	UpdateTask(runningID, StatusProcessing, 50, "", "")

	mu.Lock()
	tasks[oldID].finishedAt = time.Now().Add(-2 * finishedTaskTTL)
	mu.Unlock()

	// Enqueueing sweeps expired finished tasks
	EnqueueTask("new", func(string) {})

	if _, ok := GetTask(oldID); ok {
		t.Error("expired finished task was not pruned")
	}
	if _, ok := GetTask(runningID); !ok {
		t.Error("in-progress task must never be pruned")
	}
}

func TestFinishedTaskKeptUntilTTL(t *testing.T) {
	resetForTest(10)
	id, _ := EnqueueTask("t", func(string) {})
	UpdateTask(id, StatusFailed, 10, "", "boom")
	EnqueueTask("another", func(string) {})
	if task, ok := GetTask(id); !ok || task.Error != "boom" {
		t.Errorf("recently finished task should remain queryable, got %+v ok=%v", task, ok)
	}
}
