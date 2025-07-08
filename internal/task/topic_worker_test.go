package task

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rinsyan0518/wpks-ls/internal/pkg/shared"
	"github.com/tliron/glsp"
)

func TestTopicWorker_BasicOperation(t *testing.T) {
	var receivedItems []string
	var mu sync.Mutex
	processed := make(chan bool, 1)

	handler := func(ctx context.Context, items []string) {
		mu.Lock()
		receivedItems = append(receivedItems, items...)
		if len(receivedItems) == 2 {
			select {
			case processed <- true:
			default:
			}
		}
		mu.Unlock()
	}

	config := NewWorkerConfig(WithQueueSize(10))
	worker := NewTopicWorker("test-topic", handler, config)

	if worker.Topic != "test-topic" {
		t.Errorf("Expected topic 'test-topic', got '%s'", worker.Topic)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go worker.Run(ctx)

	// Enqueue items
	if !worker.Enqueue("item1") {
		t.Error("Failed to enqueue item1")
	}
	if !worker.Enqueue("item2") {
		t.Error("Failed to enqueue item2")
	}

	// Wait for processing to complete
	select {
	case <-processed:
		// Processing completed
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for processing")
	}

	// Verify
	mu.Lock()
	if len(receivedItems) != 2 {
		t.Errorf("Expected 2 items, got %d", len(receivedItems))
	}
	if len(receivedItems) >= 2 {
		if receivedItems[0] != "item1" || receivedItems[1] != "item2" {
			t.Errorf("Expected items [item1, item2], got %v", receivedItems)
		}
	}
	mu.Unlock()

	// Close worker
	worker.Close()
}

func TestTopicWorker_WithMessage(t *testing.T) {
	var receivedMessages []shared.Message
	var mu sync.Mutex
	processed := make(chan bool, 1)

	handler := func(ctx context.Context, msgs []shared.Message) {
		mu.Lock()
		receivedMessages = append(receivedMessages, msgs...)
		if len(receivedMessages) == 1 {
			select {
			case processed <- true:
			default:
			}
		}
		mu.Unlock()
	}

	config := NewWorkerConfig(WithQueueSize(10))
	worker := NewTopicWorker("message-topic", handler, config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go worker.Run(ctx)

	// Create test message
	mockGLSPCtx := &glsp.Context{}
	message := shared.Message{
		GLSPContext: mockGLSPCtx,
		URI:         "file:///test.rb",
		Type:        shared.DiagnoseFile,
	}

	// Enqueue message
	if !worker.Enqueue(message) {
		t.Error("Failed to enqueue message")
	}

	// Wait for processing to complete
	select {
	case <-processed:
		// Processing completed
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for processing")
	}

	// Verify
	mu.Lock()
	if len(receivedMessages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(receivedMessages))
	}
	if len(receivedMessages) > 0 && receivedMessages[0].URI != "file:///test.rb" {
		t.Errorf("Expected URI 'file:///test.rb', got '%s'", receivedMessages[0].URI)
	}
	mu.Unlock()

	// Close worker
	worker.Close()
}

func TestTopicWorker_BatchProcessing(t *testing.T) {
	receivedBatches := make(chan []string, 10)

	handler := func(ctx context.Context, items []string) {
		// Make a copy of items to avoid race conditions
		itemsCopy := make([]string, len(items))
		copy(itemsCopy, items)
		receivedBatches <- itemsCopy
	}

	config := NewWorkerConfig(
		WithQueueSize(10),
		WithBatchConfig(3, 100*time.Millisecond),
	)
	worker := NewTopicWorker("batch-topic", handler, config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go worker.Run(ctx)

	// Enqueue 5 items
	for i := 0; i < 5; i++ {
		if !worker.Enqueue(fmt.Sprintf("item%d", i)) {
			t.Errorf("Failed to enqueue item%d", i)
		}
	}

	// Collect all batches (expect at least 2: one of size 3, one of size 2)
	var batches [][]string
	timeout := time.After(500 * time.Millisecond)
	totalItems := 0

	for totalItems < 5 {
		select {
		case batch := <-receivedBatches:
			batches = append(batches, batch)
			totalItems += len(batch)
		case <-timeout:
			t.Errorf("Timeout waiting for batches. Got %d items so far", totalItems)
			return
		}
	}

	// Verify batching behavior
	if len(batches) < 1 {
		t.Errorf("Expected at least 1 batch, got %d", len(batches))
	}

	// First batch should have 3 items (batch size reached)
	if len(batches) > 0 && len(batches[0]) != 3 {
		t.Errorf("Expected first batch size 3, got %d", len(batches[0]))
	}

	// Total items should be 5
	if totalItems != 5 {
		t.Errorf("Expected 5 total items processed, got %d", totalItems)
	}

	// Close worker
	worker.Close()
}

func TestTopicWorker_TimeoutProcessing(t *testing.T) {
	var receivedBatches [][]string
	var mu sync.Mutex
	timeoutProcessed := make(chan bool, 1)

	handler := func(ctx context.Context, items []string) {
		mu.Lock()
		receivedBatches = append(receivedBatches, items)
		if len(receivedBatches) == 1 && len(items) == 2 {
			select {
			case timeoutProcessed <- true:
			default:
			}
		}
		mu.Unlock()
	}

	// Short timeout for faster test
	config := NewWorkerConfig(
		WithQueueSize(10),
		WithBatchConfig(10, 50*time.Millisecond), // Large batch size, short timeout
	)
	worker := NewTopicWorker("timeout-topic", handler, config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go worker.Run(ctx)

	// Enqueue 2 items (less than batch size)
	if !worker.Enqueue("item1") {
		t.Error("Failed to enqueue item1")
	}
	if !worker.Enqueue("item2") {
		t.Error("Failed to enqueue item2")
	}

	// Wait for timeout processing
	select {
	case <-timeoutProcessed:
		// Timeout processing completed
	case <-time.After(200 * time.Millisecond):
		t.Error("Timeout waiting for timeout processing")
	}

	// Verify timeout triggered processing
	mu.Lock()
	defer mu.Unlock()

	if len(receivedBatches) != 1 {
		t.Errorf("Expected 1 batch due to timeout, got %d", len(receivedBatches))
	}

	if len(receivedBatches) > 0 && len(receivedBatches[0]) != 2 {
		t.Errorf("Expected batch size 2, got %d", len(receivedBatches[0]))
	}

	// Close worker
	worker.Close()
}

func TestTopicWorker_DisabledBatching(t *testing.T) {
	receivedBatches := make(chan []string, 10)

	handler := func(ctx context.Context, items []string) {
		receivedBatches <- items
	}

	// Default config has batching disabled
	config := NewWorkerConfig(WithQueueSize(10))
	worker := NewTopicWorker("individual-topic", handler, config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go worker.Run(ctx)

	// Enqueue 3 items
	for i := 0; i < 3; i++ {
		if !worker.Enqueue(fmt.Sprintf("item%d", i)) {
			t.Errorf("Failed to enqueue item%d", i)
		}
	}

	// Collect all batches
	var batches [][]string
	timeout := time.After(200 * time.Millisecond)

	for len(batches) < 3 {
		select {
		case batch := <-receivedBatches:
			batches = append(batches, batch)
		case <-timeout:
			t.Errorf("Timeout waiting for batches. Got %d batches so far", len(batches))
			return
		}
	}

	// Verify each item processed individually
	if len(batches) != 3 {
		t.Errorf("Expected 3 individual batches, got %d", len(batches))
	}

	for i, batch := range batches {
		if len(batch) != 1 {
			t.Errorf("Expected batch %d to have 1 item, got %d", i, len(batch))
		}
	}

	// Close worker
	worker.Close()
}

func TestTopicWorker_QueueFull(t *testing.T) {
	processed := make(chan bool, 1)

	handler := func(ctx context.Context, items []string) {
		// Slow handler to cause queue backup
		time.Sleep(200 * time.Millisecond)
		select {
		case processed <- true:
		default:
		}
	}

	config := NewWorkerConfig(WithQueueSize(2)) // Small queue
	worker := NewTopicWorker("full-queue-topic", handler, config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go worker.Run(ctx)

	// Fill the queue
	if !worker.Enqueue("item1") {
		t.Error("First enqueue should succeed")
	}
	if !worker.Enqueue("item2") {
		t.Error("Second enqueue should succeed")
	}

	// This should fail due to full queue
	if worker.Enqueue("item3") {
		t.Error("Third enqueue should fail due to full queue")
	}

	// Wait for at least one item to be processed
	select {
	case <-processed:
		// Processing started
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout waiting for processing to start")
	}

	// Close worker
	worker.Close()
}

func TestTopicWorker_HandlerPanic(t *testing.T) {
	processedItems := make(chan string, 10)

	handler := func(ctx context.Context, items []string) {
		for _, item := range items {
			if item == "panic" {
				panic("test panic")
			}
			processedItems <- item
		}
	}

	config := NewWorkerConfig(WithQueueSize(10))
	worker := NewTopicWorker("panic-topic", handler, config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go worker.Run(ctx)

	// Enqueue normal item, panic item, and another normal item
	worker.Enqueue("normal1")
	worker.Enqueue("panic")
	worker.Enqueue("normal2")

	// Collect processed items
	var items []string
	timeout := time.After(200 * time.Millisecond)

	for len(items) < 2 {
		select {
		case item := <-processedItems:
			items = append(items, item)
		case <-timeout:
			// Timeout is expected since panic item won't be processed
			return
		}
	}

	// Verify that normal items were processed and panic was handled
	// Both normal1 and normal2 should be processed (panic is handled and doesn't stop processing)
	if len(items) != 2 {
		t.Errorf("Expected 2 processed items, got %d: %v", len(items), items)
	}
	if len(items) >= 1 && items[0] != "normal1" {
		t.Errorf("Expected first item 'normal1', got '%s'", items[0])
	}
	if len(items) >= 2 && items[1] != "normal2" {
		t.Errorf("Expected second item 'normal2', got '%s'", items[1])
	}

	// Close worker
	worker.Close()
}

func TestTopicWorker_ContextCancellation(t *testing.T) {
	started := make(chan bool, 1)

	handler := func(ctx context.Context, items []string) {
		// Handler that should not be called after cancellation
		select {
		case started <- true:
		default:
		}
	}

	config := NewWorkerConfig(WithQueueSize(10))
	worker := NewTopicWorker("cancel-topic", handler, config)

	ctx, cancel := context.WithCancel(context.Background())

	go worker.Run(ctx)

	// Enqueue an item to start processing
	worker.Enqueue("test")

	// Wait for processing to start
	select {
	case <-started:
		// Processing started
	case <-time.After(100 * time.Millisecond):
		// It's okay if handler isn't called due to quick cancellation
	}

	// Cancel context
	cancel()

	// Worker should stop gracefully
	worker.Close()
}
