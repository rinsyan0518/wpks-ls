package task

import (
	"sync"
	"testing"
	"time"
)

// Simple test struct for testing generic functionality
type testMessage struct {
	Topic string
	Value string
}

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

func TestRingBuffer_WithStructType(t *testing.T) {
	// Test with struct type to verify generic functionality
	rb := NewRingBuffer[testMessage](5)

	msg1 := testMessage{
		Topic: "test-topic",
		Value: "test-value-1",
	}

	msg2 := testMessage{
		Topic: "another-topic",
		Value: "test-value-2",
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
	if !ok || receivedMsg1.Topic != "test-topic" {
		t.Errorf("Expected 'test-topic', got '%s'", receivedMsg1.Topic)
	}

	receivedMsg2, ok := rb.Get()
	if !ok || receivedMsg2.Topic != "another-topic" {
		t.Errorf("Expected 'another-topic', got '%s'", receivedMsg2.Topic)
	}
}

func TestRingBuffer_SizeAndCapacity(t *testing.T) {
	rb := NewRingBuffer[int](5)

	// Test capacity
	if rb.Capacity() != 5 {
		t.Errorf("Expected capacity 5, got %d", rb.Capacity())
	}

	// Test size progression
	if rb.Size() != 0 {
		t.Errorf("Expected size 0, got %d", rb.Size())
	}

	rb.Put(1)
	if rb.Size() != 1 {
		t.Errorf("Expected size 1, got %d", rb.Size())
	}

	rb.Put(2)
	if rb.Size() != 2 {
		t.Errorf("Expected size 2, got %d", rb.Size())
	}

	rb.Get()
	if rb.Size() != 1 {
		t.Errorf("Expected size 1 after get, got %d", rb.Size())
	}
}
