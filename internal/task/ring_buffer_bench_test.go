package task

import (
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

func BenchmarkRingBuffer_StructType(b *testing.B) {
	type testData struct {
		ID    int
		Value string
	}

	rb := NewRingBuffer[testData](1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := testData{ID: i, Value: "test"}
		rb.TryPut(data)
		rb.TryGet()
	}
}
