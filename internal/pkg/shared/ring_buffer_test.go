package shared

import (
	"sync"
	"testing"
	"time"
)

func TestRingBuffer_BasicOperations(t *testing.T) {
	rb := NewRingBuffer[int](3)

	// Test Put and Get
	if !rb.Put(1) {
		t.Error("Put should succeed")
	}
	if !rb.Put(2) {
		t.Error("Put should succeed")
	}
	if !rb.Put(3) {
		t.Error("Put should succeed")
	}

	// Buffer should be full now
	if !rb.IsFull() {
		t.Error("Buffer should be full")
	}

	// Get items
	item, ok := rb.Get()
	if !ok || item != 1 {
		t.Errorf("Expected 1, got %d", item)
	}

	item, ok = rb.Get()
	if !ok || item != 2 {
		t.Errorf("Expected 2, got %d", item)
	}

	item, ok = rb.Get()
	if !ok || item != 3 {
		t.Errorf("Expected 3, got %d", item)
	}

	// Buffer should be empty now
	if !rb.IsEmpty() {
		t.Error("Buffer should be empty")
	}
}

func TestRingBuffer_TryOperations(t *testing.T) {
	rb := NewRingBuffer[string](2)

	// TryPut when not full
	if !rb.TryPut("a") {
		t.Error("TryPut should succeed")
	}
	if !rb.TryPut("b") {
		t.Error("TryPut should succeed")
	}

	// TryPut when full
	if rb.TryPut("c") {
		t.Error("TryPut should fail when full")
	}

	// TryGet when not empty
	item, ok := rb.TryGet()
	if !ok || item != "a" {
		t.Errorf("Expected 'a', got '%s'", item)
	}

	item, ok = rb.TryGet()
	if !ok || item != "b" {
		t.Errorf("Expected 'b', got '%s'", item)
	}

	// TryGet when empty
	_, ok = rb.TryGet()
	if ok {
		t.Error("TryGet should fail when empty")
	}
}

func TestRingBuffer_ConcurrentAccess(t *testing.T) {
	rb := NewRingBuffer[int](100)
	var wg sync.WaitGroup

	// Producer goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				rb.Put(start*100 + j)
			}
		}(i)
	}

	// Consumer goroutines
	received := make([]int, 0, 100)
	var mu sync.Mutex
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				item, ok := rb.Get()
				if !ok {
					break
				}
				mu.Lock()
				received = append(received, item)
				mu.Unlock()
			}
		}()
	}

	// Wait for producers to finish
	go func() {
		time.Sleep(100 * time.Millisecond)
		rb.Close()
	}()

	wg.Wait()

	// Verify all items received
	if len(received) != 100 {
		t.Errorf("Expected 100 items, got %d", len(received))
	}
}

func TestRingBuffer_CloseOperation(t *testing.T) {
	rb := NewRingBuffer[int](3)

	// Put some items
	rb.Put(1)
	rb.Put(2)

	// Close the buffer
	rb.Close()

	// Put should fail after close
	if rb.Put(3) {
		t.Error("Put should fail after close")
	}

	// Get should still work for existing items
	item, ok := rb.Get()
	if !ok || item != 1 {
		t.Errorf("Expected 1, got %d", item)
	}

	item, ok = rb.Get()
	if !ok || item != 2 {
		t.Errorf("Expected 2, got %d", item)
	}

	// Get should fail when empty and closed
	_, ok = rb.Get()
	if ok {
		t.Error("Get should fail when empty and closed")
	}
}

func TestRingBuffer_WithTopicMessage(t *testing.T) {
	// Test with actual topicMessage type
	rb := NewRingBuffer[topicMessage](5)

	msg1 := topicMessage{
		topic:   "test-topic",
		message: Message{URI: "file:///test.rb"},
	}

	msg2 := topicMessage{
		topic:   "another-topic",
		message: Message{URI: "file:///another.rb"},
	}

	// Put messages
	if !rb.Put(msg1) {
		t.Error("Put should succeed")
	}
	if !rb.Put(msg2) {
		t.Error("Put should succeed")
	}

	// Get messages
	receivedMsg1, ok := rb.Get()
	if !ok || receivedMsg1.topic != "test-topic" {
		t.Errorf("Expected 'test-topic', got '%s'", receivedMsg1.topic)
	}

	receivedMsg2, ok := rb.Get()
	if !ok || receivedMsg2.topic != "another-topic" {
		t.Errorf("Expected 'another-topic', got '%s'", receivedMsg2.topic)
	}
}
