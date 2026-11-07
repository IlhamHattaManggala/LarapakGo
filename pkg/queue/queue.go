package queue

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Queue defines the standard interface for LarapakGo queue drivers.
type Queue interface {
	Push(job Job) error
	Pop() (Job, bool)
	Work(stopCh <-chan struct{})
}

var (
	defaultQueue Queue
	queueOnce    sync.Once
)

// Default returns the active Queue driver based on QUEUE_DRIVER env variable.
// Defaults to in-memory queue, or RedisQueue if set to "redis".
func Default() Queue {
	queueOnce.Do(func() {
		driver := os.Getenv("QUEUE_DRIVER")
		if driver == "redis" {
			defaultQueue = NewRedisQueueFromEnv()
		} else {
			defaultQueue = NewMemoryQueue()
		}
	})
	return defaultQueue
}

// SetDefaultQueue manually sets the global queue driver.
func SetDefaultQueue(q Queue) {
	defaultQueue = q
}

// Job defines the interface for executable background jobs in LarapakGo.
type Job interface {
	Handle() error
}

// MemoryQueue provides in-memory thread-safe queue dispatching.
type MemoryQueue struct {
	jobs []Job
	mu   sync.Mutex
}

// NewMemoryQueue creates a new MemoryQueue.
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		jobs: make([]Job, 0),
	}
}

// Push enqueues a new background job.
func (q *MemoryQueue) Push(job Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, job)
	return nil
}

// Pop dequeues the oldest job.
func (q *MemoryQueue) Pop() (Job, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.jobs) == 0 {
		return nil, false
	}
	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	return job, true
}

// Work runs a continuous worker processing queued jobs.
func (q *MemoryQueue) Work(stopCh <-chan struct{}) {
	log.Println("[Queue Worker] Worker aktif, nunggu antrean gawean...")

	for {
		select {
		case <-stopCh:
			log.Println("[Queue Worker] Worker dipateni.")
			return
		default:
			job, ok := q.Pop()
			if ok {
				log.Println("[Queue Worker] Memproses pekerjaan...")
				start := time.Now()
				if err := job.Handle(); err != nil {
					log.Printf("[Queue Worker] Gawean kliru/gagal: %v", err)
				} else {
					log.Printf("[Queue Worker] Gawean sukses diselesaikna (%v).", time.Since(start))
				}
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

// FunctionJob wraps an anonymous function as an executable Job.
type FunctionJob struct {
	Name string
	Fn   func() error
}

// Handle executes the wrapped function.
func (f *FunctionJob) Handle() error {
	if f.Fn == nil {
		return fmt.Errorf("job function is nil")
	}
	return f.Fn()
}

