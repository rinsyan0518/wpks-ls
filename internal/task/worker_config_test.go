package task

import (
	"testing"
	"time"
)

func TestNewWorkerConfig_DefaultValues(t *testing.T) {
	config := NewWorkerConfig()

	if config.QueueSize != 100 {
		t.Errorf("Expected default QueueSize 100, got %d", config.QueueSize)
	}
	if config.BatchSize != 1 {
		t.Errorf("Expected default BatchSize 1, got %d", config.BatchSize)
	}
	if config.BatchTimeout != 10*time.Millisecond {
		t.Errorf("Expected default BatchTimeout 10ms, got %v", config.BatchTimeout)
	}
	if config.Enabled != false {
		t.Errorf("Expected default Enabled false, got %v", config.Enabled)
	}
}

func TestWithQueueSize(t *testing.T) {
	config := NewWorkerConfig(WithQueueSize(50))

	if config.QueueSize != 50 {
		t.Errorf("Expected QueueSize 50, got %d", config.QueueSize)
	}
	// Other values should remain default
	if config.BatchSize != 1 {
		t.Errorf("Expected default BatchSize 1, got %d", config.BatchSize)
	}
}

func TestWithBatchConfig(t *testing.T) {
	config := NewWorkerConfig(WithBatchConfig(5, 200*time.Millisecond))

	if config.BatchSize != 5 {
		t.Errorf("Expected BatchSize 5, got %d", config.BatchSize)
	}
	if config.BatchTimeout != 200*time.Millisecond {
		t.Errorf("Expected BatchTimeout 200ms, got %v", config.BatchTimeout)
	}
	if config.Enabled != true {
		t.Errorf("Expected Enabled true, got %v", config.Enabled)
	}
	// QueueSize should remain default
	if config.QueueSize != 100 {
		t.Errorf("Expected default QueueSize 100, got %d", config.QueueSize)
	}
}

func TestMultipleOptions(t *testing.T) {
	config := NewWorkerConfig(
		WithQueueSize(200),
		WithBatchConfig(10, 500*time.Millisecond),
	)

	if config.QueueSize != 200 {
		t.Errorf("Expected QueueSize 200, got %d", config.QueueSize)
	}
	if config.BatchSize != 10 {
		t.Errorf("Expected BatchSize 10, got %d", config.BatchSize)
	}
	if config.BatchTimeout != 500*time.Millisecond {
		t.Errorf("Expected BatchTimeout 500ms, got %v", config.BatchTimeout)
	}
	if config.Enabled != true {
		t.Errorf("Expected Enabled true, got %v", config.Enabled)
	}
}

func TestOptionOrder(t *testing.T) {
	// Test that later options override earlier ones
	config := NewWorkerConfig(
		WithQueueSize(100),
		WithQueueSize(200), // This should override the previous value
	)

	if config.QueueSize != 200 {
		t.Errorf("Expected QueueSize 200 (last option should win), got %d", config.QueueSize)
	}
}
