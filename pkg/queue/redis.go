package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IlhamHattaManggala/larapak-go/pkg/redis"
)

// SerializableJob allows jobs to define their unique identifier and serialized payload.
type SerializableJob interface {
	Job
	JobName() string
}

type jobPayloadWrapper struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]func() Job)
)

// RegisterJob binds a unique job name to a factory returning a fresh Job instance.
func RegisterJob(name string, factory func() Job) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[name] = factory
}

// RedisQueue provides distributed, persistent queue processing backed by Redis list operations.
type RedisQueue struct {
	client   *redis.Client
	queueKey string
}

// NewRedisQueue initializes a RedisQueue with the given Redis client and queue key name.
func NewRedisQueue(client *redis.Client, queueKey ...string) *RedisQueue {
	key := "larapak:queue:default"
	if len(queueKey) > 0 && queueKey[0] != "" {
		key = queueKey[0]
	}
	return &RedisQueue{
		client:   client,
		queueKey: key,
	}
}

// NewRedisQueueFromEnv initializes a RedisQueue using environment variables.
func NewRedisQueueFromEnv() *RedisQueue {
	return NewRedisQueue(redis.NewFromEnv())
}

// Push enqueues a job into Redis using LPUSH.
func (q *RedisQueue) Push(job Job) error {
	if q.client == nil {
		return errors.New("redis queue: client is nil")
	}

	name := "generic"
	var rawData []byte
	var err error

	if sj, ok := job.(SerializableJob); ok {
		name = sj.JobName()
		rawData, err = json.Marshal(sj)
		if err != nil {
			return fmt.Errorf("redis queue: failed to marshal job: %w", err)
		}
	} else if fj, ok := job.(*FunctionJob); ok {
		name = fj.Name
		if name == "" {
			name = "anonymous"
		}
	} else {
		rawData, err = json.Marshal(job)
		if err != nil {
			return fmt.Errorf("redis queue: failed to marshal job: %w", err)
		}
	}

	wrapper := jobPayloadWrapper{
		Name: name,
		Data: rawData,
	}

	encoded, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("redis queue: failed to wrap job payload: %w", err)
	}

	_, err = q.client.LPush(q.queueKey, string(encoded))
	return err
}

// Pop dequeues the oldest job from Redis using RPOP.
func (q *RedisQueue) Pop() (Job, bool) {
	if q.client == nil {
		return nil, false
	}

	data, err := q.client.RPop(q.queueKey)
	if err != nil || data == "" {
		return nil, false
	}

	job, err := unmarshalJob(data)
	if err != nil {
		return nil, false
	}

	return job, true
}

func unmarshalJob(data string) (Job, error) {
	var wrapper jobPayloadWrapper
	if err := json.Unmarshal([]byte(data), &wrapper); err != nil {
		return nil, err
	}

	registryMu.RLock()
	factory, exists := registry[wrapper.Name]
	registryMu.RUnlock()

	if exists && factory != nil {
		freshJob := factory()
		if len(wrapper.Data) > 0 {
			_ = json.Unmarshal(wrapper.Data, freshJob)
		}
		return freshJob, nil
	}

	// Fallback generic executable job wrapper
	return &FunctionJob{
		Name: wrapper.Name,
		Fn: func() error {
			log.Printf("[Redis Queue] Menjalankan generic job: %s", wrapper.Name)
			return nil
		},
	}, nil
}

// Work runs a non-busy worker loop using BRPOP for low-CPU event-driven queue dispatching.
func (q *RedisQueue) Work(stopCh <-chan struct{}) {
	log.Printf("[Redis Queue] Worker aktif neng list '%s', nunggu antrean gawean...", q.queueKey)

	for {
		select {
		case <-stopCh:
			log.Println("[Redis Queue] Worker dipateni.")
			return
		default:
			// BRPOP blocks up to 2 seconds waiting for incoming jobs
			results, err := q.client.BRPop(2*time.Second, q.queueKey)
			if err != nil {
				if errors.Is(err, redis.ErrNil) {
					continue
				}
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if len(results) >= 2 {
				payload := results[1]
				job, err := unmarshalJob(payload)
				if err != nil {
					log.Printf("[Redis Queue] Gagal unmarshal gawean: %v", err)
					continue
				}

				start := time.Now()
				log.Println("[Redis Queue] Memproses pekerjaan...")
				if err := job.Handle(); err != nil {
					log.Printf("[Redis Queue] Gawean kliru/gagal: %v", err)
				} else {
					log.Printf("[Redis Queue] Gawean sukses diselesaikna (%v).", time.Since(start))
				}
			}
		}
	}
}

