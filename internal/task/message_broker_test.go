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

func TestMessageBroker_BasicOperation(t *testing.T) {
	var receivedMessages []shared.Message
	var mu sync.Mutex

	handler := func(ctx context.Context, msgs []shared.Message) {
		mu.Lock()
		receivedMessages = append(receivedMessages, msgs...)
		mu.Unlock()
	}

	broker := NewMessageBroker()
	broker.RegisterTopic("test-topic", handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker.Start(ctx)
	defer broker.Close()

	// Create test message
	mockGLSPCtx := &glsp.Context{}
	message := shared.Message{
		GLSPContext: mockGLSPCtx,
		URI:         "file:///test.rb",
		Type:        shared.DiagnoseFile,
	}

	broker.Enqueue("test-topic", message)

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Verify
	mu.Lock()
	if len(receivedMessages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(receivedMessages))
	}
	if len(receivedMessages) > 0 && receivedMessages[0].URI != "file:///test.rb" {
		t.Errorf("Expected URI 'file:///test.rb', got '%s'", receivedMessages[0].URI)
	}
	mu.Unlock()
}

func TestMessageBroker_RegisterTopic_DuplicateError(t *testing.T) {
	handler := func(ctx context.Context, msgs []shared.Message) {}

	broker := NewMessageBroker()
	broker.RegisterTopic("duplicate-topic", handler)

	// Should panic on duplicate registration
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for duplicate topic registration")
		} else if fmt.Sprint(r) != "topic already registered: duplicate-topic" {
			t.Errorf("Expected specific panic message, got: %v", r)
		}
	}()

	broker.RegisterTopic("duplicate-topic", handler)
}

func TestMessageBroker_RegisterTopic_AfterStartError(t *testing.T) {
	handler1 := func(ctx context.Context, msgs []shared.Message) {}
	handler2 := func(ctx context.Context, msgs []shared.Message) {}

	broker := NewMessageBroker()
	broker.RegisterTopic("topic1", handler1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker.Start(ctx)
	defer broker.Close()

	// Should panic when registering after start
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for registration after start")
		}
	}()

	broker.RegisterTopic("topic2", handler2)
}

func TestMessageBroker_Start_NoTopicsError(t *testing.T) {
	broker := NewMessageBroker()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Should panic when starting with no topics
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for start with no topics")
		}
	}()

	broker.Start(ctx)
}

func TestMessageBroker_Enqueue_NotRunningError(t *testing.T) {
	handler := func(ctx context.Context, msgs []shared.Message) {}
	broker := NewMessageBroker()
	broker.RegisterTopic("test-topic", handler)

	message := shared.Message{URI: "file:///test.rb", Type: shared.DiagnoseFile}

	// Should panic when enqueuing before start
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for enqueue before start")
		}
	}()

	broker.Enqueue("test-topic", message)
}

func TestMessageBroker_Enqueue_TopicNotRegisteredError(t *testing.T) {
	handler := func(ctx context.Context, msgs []shared.Message) {}
	broker := NewMessageBroker()
	broker.RegisterTopic("registered-topic", handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker.Start(ctx)
	defer broker.Close()

	message := shared.Message{URI: "file:///test.rb", Type: shared.DiagnoseFile}

	// Should panic when enqueuing to unregistered topic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for unregistered topic")
		}
	}()

	broker.Enqueue("unregistered-topic", message)
}

func TestMessageBroker_MultipleTopics(t *testing.T) {
	var topicAMessages []shared.Message
	var topicBMessages []shared.Message
	var mu sync.Mutex

	handlerA := func(ctx context.Context, msgs []shared.Message) {
		mu.Lock()
		topicAMessages = append(topicAMessages, msgs...)
		mu.Unlock()
	}

	handlerB := func(ctx context.Context, msgs []shared.Message) {
		mu.Lock()
		topicBMessages = append(topicBMessages, msgs...)
		mu.Unlock()
	}

	broker := NewMessageBroker()
	broker.RegisterTopic("topic-a", handlerA)
	broker.RegisterTopic("topic-b", handlerB)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker.Start(ctx)
	defer broker.Close()

	// Enqueue to different topics
	messageA := shared.Message{URI: "file:///testA.rb", Type: shared.DiagnoseFile}
	messageB := shared.Message{URI: "file:///testB.rb", Type: shared.DiagnoseAll}

	broker.Enqueue("topic-a", messageA)
	broker.Enqueue("topic-b", messageB)

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Verify
	mu.Lock()
	defer mu.Unlock()

	if len(topicAMessages) != 1 {
		t.Errorf("Expected 1 message for topic-a, got %d", len(topicAMessages))
	}
	if len(topicBMessages) != 1 {
		t.Errorf("Expected 1 message for topic-b, got %d", len(topicBMessages))
	}

	if len(topicAMessages) > 0 && topicAMessages[0].URI != "file:///testA.rb" {
		t.Errorf("Expected topic-a URI 'file:///testA.rb', got '%s'", topicAMessages[0].URI)
	}
	if len(topicBMessages) > 0 && topicBMessages[0].URI != "file:///testB.rb" {
		t.Errorf("Expected topic-b URI 'file:///testB.rb', got '%s'", topicBMessages[0].URI)
	}
}

func TestMessageBroker_WithBatchConfig(t *testing.T) {
	var receivedBatches [][]shared.Message
	var mu sync.Mutex

	handler := func(ctx context.Context, msgs []shared.Message) {
		mu.Lock()
		receivedBatches = append(receivedBatches, msgs)
		mu.Unlock()
	}

	broker := NewMessageBroker()
	broker.RegisterTopic("batch-topic", handler,
		WithQueueSize(50),
		WithBatchConfig(3, 100*time.Millisecond),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker.Start(ctx)
	defer broker.Close()

	// Enqueue 5 messages
	for i := 0; i < 5; i++ {
		message := shared.Message{
			URI:  fmt.Sprintf("file:///test%d.rb", i),
			Type: shared.DiagnoseFile,
		}
		broker.Enqueue("batch-topic", message)
	}

	// Wait for processing
	time.Sleep(150 * time.Millisecond)

	// Verify batching
	mu.Lock()
	defer mu.Unlock()

	// Should have at least one batch
	if len(receivedBatches) < 1 {
		t.Errorf("Expected at least 1 batch, got %d", len(receivedBatches))
	}

	// First batch should have 3 messages
	if len(receivedBatches) > 0 && len(receivedBatches[0]) != 3 {
		t.Errorf("Expected first batch size 3, got %d", len(receivedBatches[0]))
	}

	// Count total processed messages
	totalMessages := 0
	for _, batch := range receivedBatches {
		totalMessages += len(batch)
	}
	if totalMessages != 5 {
		t.Errorf("Expected 5 total messages processed, got %d", totalMessages)
	}
}

func TestMessageBroker_WithQueueSizeConfig(t *testing.T) {
	handler := func(ctx context.Context, msgs []shared.Message) {
		// Slow handler to test queue overflow
		time.Sleep(200 * time.Millisecond)
	}

	broker := NewMessageBroker()
	broker.RegisterTopic("small-queue-topic", handler, WithQueueSize(2))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker.Start(ctx)
	defer broker.Close()

	message := shared.Message{URI: "file:///test.rb", Type: shared.DiagnoseFile}

	// Should succeed for first 2 messages
	broker.Enqueue("small-queue-topic", message)
	broker.Enqueue("small-queue-topic", message)

	// Third message might cause queue full panic
	defer func() {
		if r := recover(); r != nil {
			// Queue full panic is expected behavior
			if fmt.Sprint(r) != "failed to enqueue message: topic worker queue is full" {
				t.Errorf("Unexpected panic: %v", r)
			}
		}
	}()

	broker.Enqueue("small-queue-topic", message)
}

func TestMessageBroker_ConcurrentEnqueue(t *testing.T) {
	var receivedMessages []shared.Message
	var mu sync.Mutex

	handler := func(ctx context.Context, msgs []shared.Message) {
		mu.Lock()
		receivedMessages = append(receivedMessages, msgs...)
		mu.Unlock()
	}

	broker := NewMessageBroker()
	broker.RegisterTopic("concurrent-topic", handler, WithQueueSize(100))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker.Start(ctx)
	defer broker.Close()

	// Concurrent enqueue from multiple goroutines
	var wg sync.WaitGroup
	numGoroutines := 10
	messagesPerGoroutine := 5

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				message := shared.Message{
					URI:  fmt.Sprintf("file:///test_%d_%d.rb", goroutineID, j),
					Type: shared.DiagnoseFile,
				}
				broker.Enqueue("concurrent-topic", message)
			}
		}(i)
	}

	wg.Wait()

	// Wait for all processing to complete
	time.Sleep(200 * time.Millisecond)

	// Verify all messages were processed
	mu.Lock()
	defer mu.Unlock()

	expectedTotal := numGoroutines * messagesPerGoroutine
	if len(receivedMessages) != expectedTotal {
		t.Errorf("Expected %d messages, got %d", expectedTotal, len(receivedMessages))
	}
}

func TestMessageBroker_Close_Idempotent(t *testing.T) {
	handler := func(ctx context.Context, msgs []shared.Message) {}

	broker := NewMessageBroker()
	broker.RegisterTopic("test-topic", handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker.Start(ctx)

	// Close multiple times should not panic
	broker.Close()
	broker.Close()
	broker.Close()
}

func TestMessageBroker_HandlerPanic(t *testing.T) {
	var processedMessages []shared.Message
	var mu sync.Mutex

	handler := func(ctx context.Context, msgs []shared.Message) {
		mu.Lock()
		defer mu.Unlock()
		for _, msg := range msgs {
			if msg.URI == "file:///panic.rb" {
				panic("test panic")
			}
			processedMessages = append(processedMessages, msg)
		}
	}

	broker := NewMessageBroker()
	broker.RegisterTopic("panic-topic", handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker.Start(ctx)
	defer broker.Close()

	// Enqueue normal message, panic message, and another normal message
	normalMsg1 := shared.Message{URI: "file:///normal1.rb", Type: shared.DiagnoseFile}
	panicMsg := shared.Message{URI: "file:///panic.rb", Type: shared.DiagnoseFile}
	normalMsg2 := shared.Message{URI: "file:///normal2.rb", Type: shared.DiagnoseFile}

	broker.Enqueue("panic-topic", normalMsg1)
	broker.Enqueue("panic-topic", panicMsg)
	broker.Enqueue("panic-topic", normalMsg2)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify that normal messages were processed and panic was handled
	mu.Lock()
	defer mu.Unlock()

	// With batching disabled (default), each message is processed individually
	// normal1 should be processed, panic message causes panic, normal2 should be processed
	if len(processedMessages) != 2 {
		t.Errorf("Expected 2 processed messages, got %d", len(processedMessages))
	}
	if len(processedMessages) >= 1 && processedMessages[0].URI != "file:///normal1.rb" {
		t.Errorf("Expected first message 'file:///normal1.rb', got '%s'", processedMessages[0].URI)
	}
	if len(processedMessages) >= 2 && processedMessages[1].URI != "file:///normal2.rb" {
		t.Errorf("Expected second message 'file:///normal2.rb', got '%s'", processedMessages[1].URI)
	}
}
