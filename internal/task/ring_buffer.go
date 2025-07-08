package task

import (
	"sync"
)

// RingBuffer is a thread-safe ring buffer implementation
type RingBuffer[T any] struct {
	buffer   []T
	capacity int
	head     int
	tail     int
	size     int
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	closed   bool
}

// NewRingBuffer creates a new ring buffer with the specified capacity
func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	if capacity <= 0 {
		panic("capacity must be positive")
	}

	rb := &RingBuffer[T]{
		buffer:   make([]T, capacity),
		capacity: capacity,
		head:     0,
		tail:     0,
		size:     0,
		closed:   false,
	}

	rb.notEmpty = sync.NewCond(&rb.mu)
	rb.notFull = sync.NewCond(&rb.mu)

	return rb
}

// Put adds an item to the ring buffer. Blocks if buffer is full.
func (rb *RingBuffer[T]) Put(item T) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	// Wait while buffer is full and not closed
	for rb.size == rb.capacity && !rb.closed {
		rb.notFull.Wait()
	}

	// Return false if closed
	if rb.closed {
		return false
	}

	// Add item to buffer
	rb.buffer[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.capacity
	rb.size++

	// Signal that buffer is not empty
	rb.notEmpty.Signal()

	return true
}

// TryPut attempts to add an item without blocking. Returns false if buffer is full or closed.
func (rb *RingBuffer[T]) TryPut(item T) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.size == rb.capacity || rb.closed {
		return false
	}

	rb.buffer[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.capacity
	rb.size++

	rb.notEmpty.Signal()

	return true
}

// Get removes and returns an item from the ring buffer. Blocks if buffer is empty.
func (rb *RingBuffer[T]) Get() (T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	// Wait while buffer is empty and not closed
	for rb.size == 0 && !rb.closed {
		rb.notEmpty.Wait()
	}

	// Return zero value if closed and empty
	if rb.closed && rb.size == 0 {
		var zero T
		return zero, false
	}

	// Get item from buffer
	item := rb.buffer[rb.head]
	var zero T
	rb.buffer[rb.head] = zero // Clear reference for GC
	rb.head = (rb.head + 1) % rb.capacity
	rb.size--

	// Signal that buffer is not full
	rb.notFull.Signal()

	return item, true
}

// TryGet attempts to get an item without blocking. Returns false if buffer is empty.
func (rb *RingBuffer[T]) TryGet() (T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.size == 0 {
		var zero T
		return zero, false
	}

	item := rb.buffer[rb.head]
	var zero T
	rb.buffer[rb.head] = zero // Clear reference for GC
	rb.head = (rb.head + 1) % rb.capacity
	rb.size--

	rb.notFull.Signal()

	return item, true
}

// Size returns the current number of items in the buffer
func (rb *RingBuffer[T]) Size() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.size
}

// IsEmpty returns true if the buffer is empty
func (rb *RingBuffer[T]) IsEmpty() bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.size == 0
}

// IsFull returns true if the buffer is full
func (rb *RingBuffer[T]) IsFull() bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.size == rb.capacity
}

// Capacity returns the capacity of the buffer
func (rb *RingBuffer[T]) Capacity() int {
	return rb.capacity
}

// Close closes the ring buffer and wakes up all waiting goroutines
func (rb *RingBuffer[T]) Close() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if !rb.closed {
		rb.closed = true
		rb.notEmpty.Broadcast()
		rb.notFull.Broadcast()
	}
}

// IsClosed returns true if the buffer is closed
func (rb *RingBuffer[T]) IsClosed() bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.closed
}
