package shared

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tliron/glsp"
)

func TestMessageSerialJobQueue_BasicOperation(t *testing.T) {
	queue := NewMessageSerialJobQueue(10)

	// Create a mock glsp.Context (note: in real usage, this would come from the LSP server)
	mockGLSPCtx := &glsp.Context{} // This might need proper initialization in real code

	// Test handler
	var receivedMessages []Message
	var mu sync.Mutex

	handler := func(ctx context.Context, msgs []Message) {
		mu.Lock()
		receivedMessages = append(receivedMessages, msgs...)
		mu.Unlock()
	}

	// Register handler
	queue.RegisterHandler("test-topic", handler)

	// Start queue
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queue.Start(ctx)
	defer queue.Close()

	// Enqueue message
	message := Message{
		GLSPContext: mockGLSPCtx,
		URI:         "file:///test/file.rb",
		Type:        DiagnoseFile,
	}

	queue.Enqueue("test-topic", message)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify
	mu.Lock()
	if len(receivedMessages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(receivedMessages))
	}
	if len(receivedMessages) > 0 && receivedMessages[0].URI != "file:///test/file.rb" {
		t.Errorf("Expected URI 'file:///test/file.rb', got '%s'", receivedMessages[0].URI)
	}
	mu.Unlock()
}

func TestMessageSerialJobQueue_MultipleSameTopicExecution(t *testing.T) {
	queue := NewMessageSerialJobQueue(10)

	var execCount int32
	var mu sync.Mutex

	handler := func(ctx context.Context, msgs []Message) {
		mu.Lock()
		execCount += int32(len(msgs))
		mu.Unlock()
		// Simulate some work
		time.Sleep(50 * time.Millisecond)
	}

	queue.RegisterHandler("same-topic", handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queue.Start(ctx)
	defer queue.Close()

	message1 := Message{URI: "file:///test/file1.rb", Type: DiagnoseFile}
	message2 := Message{URI: "file:///test/file2.rb", Type: DiagnoseFile}

	// Enqueue same topic multiple times - all should be executed now
	queue.Enqueue("same-topic", message1)
	queue.Enqueue("same-topic", message2)
	queue.Enqueue("same-topic", message1)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify all executions (no duplicate prevention)
	mu.Lock()
	if execCount != 3 {
		t.Errorf("Expected 3 executions, got %d", execCount)
	}
	mu.Unlock()
}

func TestMessageSerialJobQueue_MultipleTopics(t *testing.T) {
	queue := NewMessageSerialJobQueue(10)

	var results sync.Map

	// Handler for topic A
	handlerA := func(ctx context.Context, msgs []Message) {
		for _, msg := range msgs {
			results.Store("topic-a", msg.URI)
		}
	}

	// Handler for topic B
	handlerB := func(ctx context.Context, msgs []Message) {
		for _, msg := range msgs {
			results.Store("topic-b", msg.URI)
		}
	}

	queue.RegisterHandler("topic-a", handlerA)
	queue.RegisterHandler("topic-b", handlerB)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queue.Start(ctx)
	defer queue.Close()

	// Enqueue different topics
	queue.Enqueue("topic-a", Message{URI: "file:///test/fileA.rb", Type: DiagnoseFile})
	queue.Enqueue("topic-b", Message{URI: "file:///test/fileB.rb", Type: DiagnoseAll})

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify both topics were processed
	valueA, okA := results.Load("topic-a")
	valueB, okB := results.Load("topic-b")

	if !okA || valueA != "file:///test/fileA.rb" {
		t.Errorf("Topic A not processed correctly")
	}
	if !okB || valueB != "file:///test/fileB.rb" {
		t.Errorf("Topic B not processed correctly")
	}
}

func TestMessageSerialJobQueue_UnregisteredTopic(t *testing.T) {
	queue := NewMessageSerialJobQueue(10)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queue.Start(ctx)
	defer queue.Close()

	// Enqueue message for unregistered topic (should not panic)
	message := Message{URI: "file:///test/file.rb", Type: DiagnoseFile}
	queue.Enqueue("unregistered-topic", message)

	// Wait a bit to ensure no crash
	time.Sleep(50 * time.Millisecond)
}

func TestMessageSerialJobQueue_BatchProcessing(t *testing.T) {
	queue := NewMessageSerialJobQueue(10)

	var receivedBatches [][]Message
	var mu sync.Mutex

	handler := func(ctx context.Context, msgs []Message) {
		mu.Lock()
		receivedBatches = append(receivedBatches, msgs)
		mu.Unlock()
	}

	// Configure batching for the topic
	queue.SetBatchConfig("batch-topic", BatchConfig{
		Size:    3,
		Enabled: true,
	})
	queue.RegisterHandler("batch-topic", handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queue.Start(ctx)
	defer queue.Close()

	// Enqueue 5 messages - only first 3 will be batched and processed
	for i := 0; i < 5; i++ {
		message := Message{
			URI:  fmt.Sprintf("file:///test/file%d.rb", i),
			Type: DiagnoseFile,
		}
		queue.Enqueue("batch-topic", message)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify batching behavior
	mu.Lock()
	defer mu.Unlock()

	// Only the first batch (3 messages) should be processed automatically
	// The remaining 2 messages won't be processed unless batch size is reached
	if len(receivedBatches) != 1 {
		t.Errorf("Expected 1 batch, got %d", len(receivedBatches))
	}

	if len(receivedBatches) >= 1 && len(receivedBatches[0]) != 3 {
		t.Errorf("Expected first batch size 3, got %d", len(receivedBatches[0]))
	}

	// Verify that batching actually worked
	totalMessages := 0
	for _, batch := range receivedBatches {
		totalMessages += len(batch)
	}
	if totalMessages != 3 {
		t.Errorf("Expected 3 total messages processed, got %d", totalMessages)
	}
}

func TestMessageSerialJobQueue_BatchDisabled(t *testing.T) {
	queue := NewMessageSerialJobQueue(10)

	var receivedBatches [][]Message
	var mu sync.Mutex

	handler := func(ctx context.Context, msgs []Message) {
		mu.Lock()
		receivedBatches = append(receivedBatches, msgs)
		mu.Unlock()
	}

	// Configure no batching for the topic
	queue.SetBatchConfig("no-batch-topic", BatchConfig{
		Size:    5,     // Large batch size
		Enabled: false, // But batching disabled
	})
	queue.RegisterHandler("no-batch-topic", handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queue.Start(ctx)
	defer queue.Close()

	// Enqueue 3 messages
	for i := 0; i < 3; i++ {
		message := Message{
			URI:  fmt.Sprintf("file:///test/file%d.rb", i),
			Type: DiagnoseFile,
		}
		queue.Enqueue("no-batch-topic", message)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify each message was processed individually
	mu.Lock()
	defer mu.Unlock()

	if len(receivedBatches) != 3 {
		t.Errorf("Expected 3 individual batches, got %d", len(receivedBatches))
	}

	for i, batch := range receivedBatches {
		if len(batch) != 1 {
			t.Errorf("Expected batch %d to have 1 message, got %d", i, len(batch))
		}
	}
}

func TestMessageSerialJobQueue_MixedTopicBatching(t *testing.T) {
	queue := NewMessageSerialJobQueue(10)

	var batchedMessages []Message
	var individualMessages []Message
	var mu sync.Mutex

	batchHandler := func(ctx context.Context, msgs []Message) {
		mu.Lock()
		batchedMessages = append(batchedMessages, msgs...)
		mu.Unlock()
	}

	individualHandler := func(ctx context.Context, msgs []Message) {
		mu.Lock()
		individualMessages = append(individualMessages, msgs...)
		mu.Unlock()
	}

	// Configure different batching for different topics
	queue.SetBatchConfig("batched", BatchConfig{Size: 2, Enabled: true})
	queue.SetBatchConfig("individual", BatchConfig{Size: 1, Enabled: false})

	queue.RegisterHandler("batched", batchHandler)
	queue.RegisterHandler("individual", individualHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queue.Start(ctx)
	defer queue.Close()

	// Enqueue messages for both topics
	for i := 0; i < 4; i++ {
		queue.Enqueue("batched", Message{URI: fmt.Sprintf("batch_%d.rb", i), Type: DiagnoseFile})
		queue.Enqueue("individual", Message{URI: fmt.Sprintf("individual_%d.rb", i), Type: DiagnoseFile})
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify results
	mu.Lock()
	defer mu.Unlock()

	if len(batchedMessages) != 4 {
		t.Errorf("Expected 4 batched messages, got %d", len(batchedMessages))
	}

	if len(individualMessages) != 4 {
		t.Errorf("Expected 4 individual messages, got %d", len(individualMessages))
	}
}
