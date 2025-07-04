package shared

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

// KeyedJobFunc is a job function with a key.
type KeyedJobFunc func(ctx context.Context)

// KeyedJobQueue is a job queue that prevents duplicate keys in the queue.
type KeyedJobQueue interface {
	Enqueue(key string, job KeyedJobFunc)
	Start(ctx context.Context)
	Close()
}

type keyedJob struct {
	key string
	job KeyedJobFunc
}

// queueState is an enum-like type for KeyedSerialJobQueue state
// stateRunning: queue is accepting and executing jobs
// stateDraining: queue is draining (after Close, queued but unstarted jobs are skipped)
// stateClosed: queue is fully closed
type queueState int32

const (
	stateRunning  queueState = iota // queue is accepting and executing jobs
	stateDraining                   // queue is draining (after Close, queued but unstarted jobs are skipped)
	stateClosed                     // queue is fully closed
)

// KeyedSerialJobQueue is a serial job queue that prevents duplicate keys in the queue.
type KeyedSerialJobQueue struct {
	queue       chan keyedJob
	pendingKeys map[string]struct{}
	mu          sync.Mutex
	state       atomic.Int32
	wg          sync.WaitGroup
	cancel      context.CancelFunc
}

func NewKeyedSerialJobQueue(buffer int) *KeyedSerialJobQueue {
	jq := &KeyedSerialJobQueue{
		queue:       make(chan keyedJob, buffer),
		pendingKeys: make(map[string]struct{}),
	}
	return jq
}

func (jq *KeyedSerialJobQueue) Start(ctx context.Context) {
	jq.state.Store(int32(stateRunning))
	ctx, cancel := context.WithCancel(ctx)
	jq.cancel = cancel
	go jq.worker(ctx)
}

// Enqueue adds a keyed job to the queue. If the key is already pending or running, it is not enqueued.
func (jq *KeyedSerialJobQueue) Enqueue(key string, job KeyedJobFunc) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	if jq.state.Load() != int32(stateRunning) {
		panic("queue not running")
	}
	if _, exists := jq.pendingKeys[key]; exists {
		return // already pending or running
	}
	jq.pendingKeys[key] = struct{}{}
	jq.wg.Add(1)
	jq.queue <- keyedJob{key, job}
}

// worker executes jobs from the queue. If draining, jobs are skipped.
func (jq *KeyedSerialJobQueue) worker(ctx context.Context) {
	for kj := range jq.queue {
		jq.mu.Lock()
		delete(jq.pendingKeys, kj.key)
		jq.mu.Unlock()
		if jq.state.Load() != int32(stateRunning) {
			jq.wg.Done()
			continue
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(os.Stderr, "job panic: %v\n", r)
				}
				jq.wg.Done()
			}()
			kj.job(ctx)
		}()
	}
}

// Close sets draining and closes the queue. Jobs remaining in the queue will not be executed.
// Waits for all running jobs to finish, then sets the state to closed.
func (jq *KeyedSerialJobQueue) Close() {
	if jq.state.Load() != int32(stateRunning) {
		return
	}

	jq.state.Store(int32(stateDraining))
	close(jq.queue)
	jq.cancel()
	jq.wg.Wait()
	jq.state.Store(int32(stateClosed))
}

var _ KeyedJobQueue = (*KeyedSerialJobQueue)(nil)
