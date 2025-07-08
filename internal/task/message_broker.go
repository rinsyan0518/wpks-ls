package task

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/shared"
)

// brokerState represents the state of the message broker
type brokerState int32

const (
	brokerStateIdle     brokerState = iota // broker is idle (not started)
	brokerStateRunning                     // broker is running
	brokerStateDraining                    // broker is draining
	brokerStateClosed                      // broker is closed
)

// Broker is a message broker that processes messages by topic.
type Broker interface {
	RegisterTopic(topic string, handler JobFunc[shared.Message], opts ...WorkerConfigOption)
	Enqueue(topic string, message shared.Message)
	Start(ctx context.Context)
	Close()
}

// MessageBroker is a topic-based job queue that processes messages using dedicated workers per topic.
type MessageBroker struct {
	workers map[string]*TopicWorker[shared.Message]
	mu      sync.RWMutex
	state   atomic.Int32
	baseCtx context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewMessageBroker creates a new topic-based message broker
func NewMessageBroker() *MessageBroker {
	return &MessageBroker{
		workers: make(map[string]*TopicWorker[shared.Message]),
	}
}

// RegisterTopic registers a topic with handler and creates worker immediately
func (b *MessageBroker) RegisterTopic(topic string, handler JobFunc[shared.Message], opts ...WorkerConfigOption) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Start worker if broker is already running
	if b.state.Load() == int32(brokerStateRunning) {
		panic(fmt.Sprintf("broker already started: %s", topic))
	}

	// Check if topic already registered
	if _, exists := b.workers[topic]; exists {
		panic(fmt.Sprintf("topic already registered: %s", topic))
	}

	// Create worker immediately
	config := NewWorkerConfig(opts...)
	worker := NewTopicWorker(topic, handler, config)
	b.workers[topic] = worker
}

// Start starts the broker and all topic workers
func (b *MessageBroker) Start(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.workers) == 0 {
		panic("no topics registered: at least one topic must be registered before starting")
	}

	b.state.Store(int32(brokerStateRunning))
	b.baseCtx, b.cancel = context.WithCancel(ctx)

	// Start all existing workers
	for _, worker := range b.workers {
		b.wg.Add(1)
		go func(w *TopicWorker[shared.Message]) {
			defer b.wg.Done()
			w.Run(b.baseCtx)
		}(worker)
	}
}

// Enqueue adds a message to the appropriate topic worker
func (b *MessageBroker) Enqueue(topic string, message shared.Message) {
	if b.state.Load() != int32(brokerStateRunning) {
		panic("broker not running")
	}

	// Get existing worker (topic must be pre-registered)
	b.mu.RLock()
	worker, exists := b.workers[topic]
	b.mu.RUnlock()
	if !exists {
		panic(fmt.Sprintf("topic not registered: %s", topic))
	}

	if !worker.Enqueue(message) {
		panic("failed to enqueue message: topic worker queue is full")
	}
}

// Close stops all topic workers and waits for them to finish
func (b *MessageBroker) Close() {
	if b.state.Load() != int32(brokerStateRunning) {
		return
	}

	b.state.Store(int32(brokerStateDraining))

	// Cancel context to stop all workers
	if b.cancel != nil {
		b.cancel()
	}

	// Close all worker queues
	b.mu.RLock()
	for _, worker := range b.workers {
		worker.Close()
	}
	b.mu.RUnlock()

	// Wait for all workers to finish
	b.wg.Wait()
	b.state.Store(int32(brokerStateClosed))
}

var _ Broker = (*MessageBroker)(nil)
