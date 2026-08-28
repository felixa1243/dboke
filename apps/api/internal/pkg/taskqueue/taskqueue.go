package taskqueue

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

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

// EnqueueTask adds a new task to the queue, returns its ID, and schedules it to run on a worker
func EnqueueTask(initialMsg string, execute func(id string)) string {
	id := GenerateID()
	mu.Lock()
	tasks[id] = &Task{
		ID:       id,
		Status:   StatusPending,
		Progress: 0,
		Message:  initialMsg,
	}
	mu.Unlock()

	// Send to worker pool
	jobQueue <- func() {
		execute(id)
	}

	return id
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
