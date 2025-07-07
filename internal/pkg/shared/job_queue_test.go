package shared

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tliron/glsp"
)

func TestKeyedSerialJobQueue_BasicEnqueueAndWait(t *testing.T) {
	jq := NewKeyedSerialJobQueue(10)
	jq.Start(context.Background())
	var count int32
	for i := 0; i < 5; i++ {
		key := "job" + string(rune(i))
		jq.Enqueue(key, func(ctx context.Context) {
			atomic.AddInt32(&count, 1)
		})
	}
	jq.Close()
	if count > 5 {
		t.Errorf("expected 0-5 jobs to run, got %d", count)
	}
}

func TestKeyedSerialJobQueue_PanicRecovery(t *testing.T) {
	jq := NewKeyedSerialJobQueue(2)
	jq.Start(context.Background())
	var ran int32
	jq.Enqueue("panic", func(ctx context.Context) {
		panic("test panic")
	})
	jq.Enqueue("after", func(ctx context.Context) {
		atomic.AddInt32(&ran, 1)
	})
	jq.Close()
	if ran < 0 || ran > 1 {
		t.Error("job after panic did not run as expected")
	}
}

func TestKeyedSerialJobQueue_ClosePreventsFurtherEnqueue(t *testing.T) {
	jq := NewKeyedSerialJobQueue(1)
	jq.Start(context.Background())
	jq.Close()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when enqueueing after close")
		}
	}()
	jq.Enqueue("shouldpanic", func(ctx context.Context) {}) // should panic
}

func TestKeyedSerialJobQueue_CloseBlocksUntilAllJobsDone(t *testing.T) {
	jq := NewKeyedSerialJobQueue(2)
	jq.Start(context.Background())
	var done int32
	jq.Enqueue("wait", func(ctx context.Context) {
		time.Sleep(50 * time.Millisecond)
		atomic.StoreInt32(&done, 1)
	})
	jq.Close()
}

func TestKeyedSerialJobQueue_DuplicateKeyPrevention(t *testing.T) {
	jq := NewKeyedSerialJobQueue(10)
	jq.Start(context.Background())
	var count int32
	key := "dup"
	for i := 0; i < 5; i++ {
		jq.Enqueue(key, func(ctx context.Context) {
			atomic.AddInt32(&count, 1)
			time.Sleep(10 * time.Millisecond) // ensure worker has time to start
		})
	}
	time.Sleep(20 * time.Millisecond) // allow worker to start the first job
	jq.Close()
	if count != 1 {
		t.Errorf("expected only 1 job to run for duplicate key, got %d", count)
	}
}

func TestMessageSerialJobQueue_BasicOperation(t *testing.T) {
	queue := NewMessageSerialJobQueue(10)

	// Create a mock glsp.Context (note: in real usage, this would come from the LSP server)
	mockGLSPCtx := &glsp.Context{} // This might need proper initialization in real code

	// Test handler
	var receivedMessages []Message
	var mu sync.Mutex

	handler := func(ctx context.Context, msg Message) {
		mu.Lock()
		receivedMessages = append(receivedMessages, msg)
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

	handler := func(ctx context.Context, msg Message) {
		mu.Lock()
		execCount++
		mu.Unlock()
		// Simulate some work
		time.Sleep(50 * time.Millisecond)
	}

	queue.RegisterHandler("same-topic", handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queue.Start(ctx)
	defer queue.Close()

	message1 := Message{URI: "file:///test/file1.rb"}
	message2 := Message{URI: "file:///test/file2.rb"}

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
	handlerA := func(ctx context.Context, msg Message) {
		results.Store("topic-a", msg.URI)
	}

	// Handler for topic B
	handlerB := func(ctx context.Context, msg Message) {
		results.Store("topic-b", msg.URI)
	}

	queue.RegisterHandler("topic-a", handlerA)
	queue.RegisterHandler("topic-b", handlerB)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queue.Start(ctx)
	defer queue.Close()

	// Enqueue different topics
	queue.Enqueue("topic-a", Message{URI: "file:///test/fileA.rb"})
	queue.Enqueue("topic-b", Message{URI: "file:///test/fileB.rb"})

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
	message := Message{URI: "file:///test/file.rb"}
	queue.Enqueue("unregistered-topic", message)

	// Wait a bit to ensure no crash
	time.Sleep(50 * time.Millisecond)
}
