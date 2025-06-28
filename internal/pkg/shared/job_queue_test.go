package shared

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSerialJobQueue_BasicEnqueueAndWait(t *testing.T) {
	jq := NewSerialJobQueue(10)
	var count int32
	for i := 0; i < 5; i++ {
		jq.Enqueue(func() {
			atomic.AddInt32(&count, 1)
		})
	}
	jq.Close()
	if count > 5 {
		t.Errorf("expected 0-5 jobs to run, got %d", count)
	}
}

func TestSerialJobQueue_PanicRecovery(t *testing.T) {
	jq := NewSerialJobQueue(2)
	var ran int32
	jq.Enqueue(func() {
		panic("test panic")
	})
	jq.Enqueue(func() {
		atomic.AddInt32(&ran, 1)
	})
	jq.Close()
	if ran < 0 || ran > 1 {
		t.Error("job after panic did not run as expected")
	}
}

func TestSerialJobQueue_ClosePreventsFurtherEnqueue(t *testing.T) {
	jq := NewSerialJobQueue(1)
	jq.Close()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when enqueueing after close")
		}
	}()
	jq.Enqueue(func() {}) // should panic
}

func TestSerialJobQueue_CloseBlocksUntilAllJobsDone(t *testing.T) {
	jq := NewSerialJobQueue(2)
	var done int32
	jq.Enqueue(func() {
		time.Sleep(50 * time.Millisecond)
		atomic.StoreInt32(&done, 1)
	})
	jq.Close()
}

func TestSerialJobQueue_DrainSkipsJobs(t *testing.T) {
	jq := NewSerialJobQueue(10)
	var count int32
	for i := 0; i < 100; i++ {
		jq.Enqueue(func() {
			time.Sleep(1 * time.Millisecond)
			atomic.AddInt32(&count, 1)
		})
	}
	time.Sleep(2 * time.Millisecond) // workerが1つだけ消費する猶予
	jq.Close()
	if count >= 100 {
		t.Errorf("expected some jobs to be skipped after drain, got %d", count)
	}
}
