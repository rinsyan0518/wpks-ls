package shared

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
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
