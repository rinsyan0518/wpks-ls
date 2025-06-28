package shared

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type JobFunc func()

type JobQueue interface {
	Enqueue(job JobFunc)
	Close()
}

// queueState is an enum-like type for SerialJobQueue state
// stateRunning: queue is accepting and executing jobs
// stateDraining: queue is draining (after Close, queued but unstarted jobs are skipped)
// stateClosed: queue is fully closed
type queueState int32

const (
	stateRunning  queueState = iota // queue is accepting and executing jobs
	stateDraining                   // queue is draining (after Close, queued but unstarted jobs are skipped)
	stateClosed                     // queue is fully closed
)

// SerialJobQueue is a channel-based serial job queue implementation with enum state.
type SerialJobQueue struct {
	queue chan JobFunc
	state atomic.Int32
	wg    sync.WaitGroup
}

func NewSerialJobQueue(buffer int) *SerialJobQueue {
	jq := &SerialJobQueue{
		queue: make(chan JobFunc, buffer),
	}
	jq.state.Store(int32(stateRunning))
	go jq.worker()
	return jq
}

// Enqueue adds a job to the queue. Panics if the queue is not running.
func (jq *SerialJobQueue) Enqueue(job JobFunc) {
	if jq.state.Load() != int32(stateRunning) {
		panic("queue not running")
	}
	jq.wg.Add(1)
	jq.queue <- job
}

// worker executes jobs from the queue. If draining, jobs are skipped.
func (jq *SerialJobQueue) worker() {
	for job := range jq.queue {
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
			job()
		}()
	}
}

// Close sets draining and closes the queue. Jobs remaining in the queue will not be executed.
// Waits for all running jobs to finish, then sets the state to closed.
func (jq *SerialJobQueue) Close() {
	jq.state.Store(int32(stateDraining))
	close(jq.queue)
	jq.wg.Wait()
	jq.state.Store(int32(stateClosed))
}

var _ JobQueue = (*SerialJobQueue)(nil)
