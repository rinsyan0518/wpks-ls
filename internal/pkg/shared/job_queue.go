package shared

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/tliron/glsp"
)

// DiagnoseType represents the type of diagnosis to perform
type DiagnoseType int

const (
	DiagnoseFile DiagnoseType = iota // Diagnose a single file
	DiagnoseAll                      // Diagnose all files
)

// Message represents a message containing glsp.Context and URI
type Message struct {
	GLSPContext *glsp.Context
	URI         string
	Type        DiagnoseType
	// Additional fields can be added here as needed
}

// MessageJobFunc is a job function that processes a message.
type MessageJobFunc func(ctx context.Context, msg Message)

// MessageJobQueue is a job queue that processes messages by topic.
type MessageJobQueue interface {
	RegisterHandler(topic string, handler MessageJobFunc)
	Enqueue(topic string, message Message)
	Start(ctx context.Context)
	Close()
}

type topicMessage struct {
	topic   string
	message Message
}

// queueState is an enum-like type for MessageSerialJobQueue state
// stateRunning: queue is accepting and executing jobs
// stateDraining: queue is draining (after Close, queued but unstarted jobs are skipped)
// stateClosed: queue is fully closed
type queueState int32

const (
	stateRunning  queueState = iota // queue is accepting and executing jobs
	stateDraining                   // queue is draining (after Close, queued but unstarted jobs are skipped)
	stateClosed                     // queue is fully closed
)

// MessageSerialJobQueue is a serial job queue that processes messages by topic.
type MessageSerialJobQueue struct {
	queue    *RingBuffer[topicMessage]
	handlers map[string]MessageJobFunc
	mu       sync.Mutex
	state    atomic.Int32
	wg       sync.WaitGroup
	cancel   context.CancelFunc
}

func NewMessageSerialJobQueue(buffer int) *MessageSerialJobQueue {
	jq := &MessageSerialJobQueue{
		queue:    NewRingBuffer[topicMessage](buffer),
		handlers: make(map[string]MessageJobFunc),
	}
	return jq
}

// RegisterHandler registers a handler function for a specific topic
func (jq *MessageSerialJobQueue) RegisterHandler(topic string, handler MessageJobFunc) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	jq.handlers[topic] = handler
}

func (jq *MessageSerialJobQueue) Start(ctx context.Context) {
	jq.state.Store(int32(stateRunning))
	ctx, cancel := context.WithCancel(ctx)
	jq.cancel = cancel
	go jq.worker(ctx)
}

// Enqueue adds a topic message to the queue.
func (jq *MessageSerialJobQueue) Enqueue(topic string, message Message) {
	if jq.state.Load() != int32(stateRunning) {
		panic("queue not running")
	}
	jq.wg.Add(1)
	if !jq.queue.Put(topicMessage{topic, message}) {
		jq.wg.Done() // Decrease counter if put failed
		panic("failed to enqueue message")
	}
}

// worker executes jobs from the queue. If draining, jobs are skipped.
func (jq *MessageSerialJobQueue) worker(ctx context.Context) {
	for {
		tm, ok := jq.queue.Get()
		if !ok {
			// Queue is closed and empty
			break
		}

		jq.mu.Lock()
		handler, exists := jq.handlers[tm.topic]
		jq.mu.Unlock()

		if jq.state.Load() != int32(stateRunning) {
			jq.wg.Done()
			continue
		}

		if !exists {
			fmt.Fprintf(os.Stderr, "no handler registered for topic: %s\n", tm.topic)
			jq.wg.Done()
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(os.Stderr, "job panic for topic %s: %v\n", tm.topic, r)
				}
				jq.wg.Done()
			}()
			handler(ctx, tm.message)
		}()
	}
}

// Close sets draining and closes the queue. Jobs remaining in the queue will not be executed.
// Waits for all running jobs to finish, then sets the state to closed.
func (jq *MessageSerialJobQueue) Close() {
	if jq.state.Load() != int32(stateRunning) {
		return
	}

	jq.state.Store(int32(stateDraining))
	jq.queue.Close()
	jq.cancel()
	jq.wg.Wait()
	jq.state.Store(int32(stateClosed))
}

var _ MessageJobQueue = (*MessageSerialJobQueue)(nil)
