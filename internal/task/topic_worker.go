package task

import (
	"context"
	"fmt"
	"os"
	"time"
)

// JobFunc is a generic job function that processes multiple items.
type JobFunc[T any] func(ctx context.Context, items []T)

// TopicWorker handles message processing for a specific topic
type TopicWorker[T any] struct {
	Topic       string
	config      *WorkerConfig
	batchBuffer *RingBuffer[T]
	handler     JobFunc[T]
	queue       chan T
	done        chan struct{}
}

// NewTopicWorker creates a new topic worker
func NewTopicWorker[T any](topic string, handler JobFunc[T], config *WorkerConfig) *TopicWorker[T] {
	return &TopicWorker[T]{
		Topic:       topic,
		queue:       make(chan T, config.QueueSize),
		batchBuffer: NewRingBuffer[T](config.BatchSize * 10),
		handler:     handler,
		config:      config,
		done:        make(chan struct{}),
	}
}

// Run starts the worker goroutine
func (w *TopicWorker[T]) Run(ctx context.Context) {
	defer close(w.done)

	var timer *time.Timer
	var timerCh <-chan time.Time

	if w.config.Enabled && w.config.BatchTimeout > 0 {
		timer = time.NewTimer(w.config.BatchTimeout)
		timer.Stop() // Stop initially
	}

	for {
		// Update timer channel
		if timer != nil {
			timerCh = timer.C
		} else {
			timerCh = nil
		}

		select {
		case msg, ok := <-w.queue:
			if !ok {
				// Queue closed, process remaining items and exit
				w.ProcessBatch(ctx)
				return
			}

			// Add message to batch buffer
			if !w.batchBuffer.TryPut(msg) {
				// Buffer full, process immediately
				w.ProcessBatch(ctx)
				// Try again after processing
				if !w.batchBuffer.TryPut(msg) {
					// Still can't put, skip this message (should not happen with proper sizing)
					continue
				}
			}

			// Check if we should process the batch now
			if w.ShouldProcessBatch() {
				if timer != nil {
					timer.Stop()
				}
				w.ProcessBatch(ctx)
			} else if timer != nil && w.config.BatchTimeout > 0 && w.batchBuffer.Size() > 0 {
				// Reset timer for timeout-based processing (only if buffer has items)
				timer.Reset(w.config.BatchTimeout)
			}

		case <-timerCh:
			// Timeout occurred, process accumulated batch
			w.ProcessBatch(ctx)

		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			// Process remaining items before exiting
			w.ProcessBatch(ctx)
			return
		}
	}
}

// ShouldProcessBatch determines if the batch should be processed immediately
func (w *TopicWorker[T]) ShouldProcessBatch() bool {
	// Don't process if buffer is empty
	if w.batchBuffer.Size() == 0 {
		return false
	}

	if !w.config.Enabled {
		return true // Process immediately if batching is disabled
	}

	return w.batchBuffer.Size() >= w.config.BatchSize
}

// ProcessBatch processes all accumulated messages in the batch buffer
func (w *TopicWorker[T]) ProcessBatch(ctx context.Context) {
	if w.batchBuffer.Size() == 0 {
		return
	}

	// Collect items from buffer
	items := make([]T, 0, w.config.BatchSize)

	// Safely drain buffer with size limit to prevent infinite loops
	bufferSize := w.batchBuffer.Size()
	for range bufferSize {
		if item, ok := w.batchBuffer.TryGet(); ok {
			items = append(items, item)
		} else {
			break // No more items
		}
	}

	if len(items) == 0 {
		return
	}

	// Process the batch
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "job panic for topic %s: %v\n", w.Topic, r)
			}
		}()
		w.handler(ctx, items)
	}()
}

// Enqueue adds an item to the worker's queue
func (w *TopicWorker[T]) Enqueue(item T) bool {
	select {
	case w.queue <- item:
		return true
	default:
		return false // Queue is full
	}
}

// Close closes the worker
func (w *TopicWorker[T]) Close() {
	close(w.queue)
	<-w.done
}
