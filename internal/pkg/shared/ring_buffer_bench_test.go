package shared

import (
	"context"
	"sync"
	"testing"
)

func BenchmarkRingBuffer_Put(b *testing.B) {
	rb := NewRingBuffer[int](1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.TryPut(i)
	}
}

func BenchmarkRingBuffer_Get(b *testing.B) {
	rb := NewRingBuffer[int](1000)

	// Fill buffer
	for i := 0; i < 1000; i++ {
		rb.Put(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.TryGet()
	}
}

func BenchmarkRingBuffer_PutGet(b *testing.B) {
	rb := NewRingBuffer[int](100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.TryPut(i)
		rb.TryGet()
	}
}

func BenchmarkRingBuffer_ConcurrentPutGet(b *testing.B) {
	rb := NewRingBuffer[int](1000)
	var wg sync.WaitGroup

	b.ResetTimer()

	// Producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < b.N; i++ {
			rb.Put(i)
		}
		rb.Close()
	}()

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, ok := rb.Get()
			if !ok {
				break
			}
		}
	}()

	wg.Wait()
}

func BenchmarkMessageJobQueue_EnqueueProcess(b *testing.B) {
	queue := NewMessageSerialJobQueue(1000)

	var processed int
	handler := func(ctx context.Context, msg Message) {
		processed++
	}

	queue.RegisterHandler("bench-topic", handler)

	ctx := context.Background()
	queue.Start(ctx)
	defer queue.Close()

	message := Message{URI: "file:///bench.rb"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Enqueue("bench-topic", message)
	}
}
