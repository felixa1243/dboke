package taskqueue

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// ErrQueueFull is returned by EnqueueTask when no worker capacity is left.
var ErrQueueFull = errors.New("task queue is full")

// finishedTaskTTL is how long completed or failed tasks stay queryable.
const finishedTaskTTL = time.Hour

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID       string     `json:"id"`
	Status   TaskStatus `json:"status"`
	Progress int        `json:"progress"`
	Message  string     `json:"message"`
	Error    string     `json:"error,omitempty"`

	finishedAt time.Time
}

var (
	tasks = make(map[string]*Task)
	mu    sync.RWMutex
	jobQueue chan func()
)

// InitWorkerPool initializes a fixed number of workers to process background tasks
func InitWorkerPool(numWorkers int) {
	jobQueue = make(chan func(), 100) // Buffer for up to 100 queued tasks
	for i := 0; i < numWorkers; i++ {
		go worker()
	}
}

func worker() {
	for job := range jobQueue {
		job() // Execute the task
	}
}

// GenerateID creates a simple random ID for a task
func GenerateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// pruneFinishedLocked drops tasks that finished longer than finishedTaskTTL ago.
// The caller must hold mu.
func pruneFinishedLocked(now time.Time) {
	for id, t := range tasks {
		if !t.finishedAt.IsZero() && now.Sub(t.finishedAt) > finishedTaskTTL {
			delete(tasks, id)
		}
	}
}

// EnqueueTask adds a new task to the queue, returns its ID, and schedules it to run on a worker.
// It never blocks: if the queue is full it returns ErrQueueFull and records nothing.
func EnqueueTask(initialMsg string, execute func(id string)) (string, error) {
	id := GenerateID()

	mu.Lock()
	pruneFinishedLocked(time.Now())
	tasks[id] = &Task{
		ID:       id,
		Status:   StatusPending,
		Progress: 0,
		Message:  initialMsg,
	}
	mu.Unlock()

	select {
	case jobQueue <- func() { execute(id) }:
		return id, nil
	default:
		mu.Lock()
		delete(tasks, id)
		mu.Unlock()
		return "", ErrQueueFull
	}
}

// GetTask returns a copy of the task
func GetTask(id string) (Task, bool) {
	mu.RLock()
	defer mu.RUnlock()
	t, ok := tasks[id]
	if !ok {
		return Task{}, false
	}
	return *t, true
}

// UpdateTask updates the progress and message of a task
func UpdateTask(id string, status TaskStatus, progress int, msg string, errMsg string) {
	mu.Lock()
	defer mu.Unlock()
	if t, ok := tasks[id]; ok {
		t.Status = status
		if (status == StatusCompleted || status == StatusFailed) && t.finishedAt.IsZero() {
			t.finishedAt = time.Now()
		}
		if progress >= 0 {
			t.Progress = progress
		}
		if msg != "" {
			t.Message = msg
		}
		if errMsg != "" {
			t.Error = errMsg
		}
	}
}
