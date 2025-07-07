package shared

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

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

// MessageJobFunc is a job function that processes multiple messages.
type MessageJobFunc func(ctx context.Context, msgs []Message)

// BatchConfig represents configuration for batch processing per topic
type BatchConfig struct {
	Size    int  // Maximum batch size
	Enabled bool // Whether batching is enabled for this topic
}

// MessageJobQueue is a job queue that processes messages by topic.
type MessageJobQueue interface {
	RegisterHandler(topic string, handler MessageJobFunc)
	SetBatchConfig(topic string, config BatchConfig)
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
	queue        *RingBuffer[topicMessage]
	handlers     map[string]MessageJobFunc
	batchConfigs map[string]BatchConfig
	mu           sync.Mutex
	state        atomic.Int32
	wg           sync.WaitGroup
	cancel       context.CancelFunc
}

func NewMessageSerialJobQueue(buffer int) *MessageSerialJobQueue {
	jq := &MessageSerialJobQueue{
		queue:        NewRingBuffer[topicMessage](buffer),
		handlers:     make(map[string]MessageJobFunc),
		batchConfigs: make(map[string]BatchConfig),
	}
	return jq
}

// SetBatchConfig sets batch configuration for a specific topic
func (jq *MessageSerialJobQueue) SetBatchConfig(topic string, config BatchConfig) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	jq.batchConfigs[topic] = config
}

// getBatchConfig returns batch configuration for a topic (with defaults)
func (jq *MessageSerialJobQueue) getBatchConfig(topic string) BatchConfig {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	if config, exists := jq.batchConfigs[topic]; exists {
		return config
	}

	// Default configuration: no batching (immediate processing)
	return BatchConfig{
		Size:    1,
		Enabled: false,
	}
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

// worker executes jobs from the queue with batch processing.
func (jq *MessageSerialJobQueue) worker(ctx context.Context) {
	batches := make(map[string][]Message) // topic -> messages

	// processAllBatches processes all accumulated batches
	processAllBatches := func() {
		for topic, messages := range batches {
			if len(messages) > 0 {
				jq.processBatch(ctx, topic, messages)
			}
		}
		batches = make(map[string][]Message) // Reset batches
	}

	for {
		select {
		case <-ctx.Done():
			// Context was cancelled, process remaining batches then exit
			processAllBatches()
			return

		default:
			// Try to get a message (non-blocking)
			tm, ok := jq.queue.TryGet()
			if !ok {
				// No message available, check if queue is closed
				if jq.queue.IsClosed() && jq.queue.IsEmpty() {
					// Queue is closed and empty, process remaining batches then exit
					processAllBatches()
					return
				}
				// No message but queue is still open, wait a bit
				time.Sleep(time.Millisecond)
				continue
			}

			// Add message to appropriate batch
			batches[tm.topic] = append(batches[tm.topic], tm.message)

			// Check if batch should be processed now
			config := jq.getBatchConfig(tm.topic)
			if !config.Enabled || len(batches[tm.topic]) >= config.Size {
				jq.processBatch(ctx, tm.topic, batches[tm.topic])
				delete(batches, tm.topic)
			}
		}
	}
}

// processBatch processes a batch of messages for a specific topic
func (jq *MessageSerialJobQueue) processBatch(ctx context.Context, topic string, messages []Message) {
	jq.mu.Lock()
	handler, exists := jq.handlers[topic]
	jq.mu.Unlock()

	if jq.state.Load() != int32(stateRunning) {
		// Mark all messages as done
		for range messages {
			jq.wg.Done()
		}
		return
	}

	if !exists {
		fmt.Fprintf(os.Stderr, "no handler registered for topic: %s\n", topic)
		// Mark all messages as done
		for range messages {
			jq.wg.Done()
		}
		return
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "job panic for topic %s: %v\n", topic, r)
			}
			// Mark all messages as done
			for range messages {
				jq.wg.Done()
			}
		}()
		handler(ctx, messages)
	}()
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
