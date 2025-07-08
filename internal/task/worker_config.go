package task

import "time"

// WorkerConfig represents configuration for a worker
type WorkerConfig struct {
	QueueSize    int           // Maximum queue size
	BatchSize    int           // Maximum batch size
	BatchTimeout time.Duration // Time interval to wait before processing batch
	Enabled      bool          // Whether batching is enabled for this topic
}

// WorkerConfigOption is a function that modifies the WorkerConfig
type WorkerConfigOption func(opts *WorkerConfig)

// NewWorkerConfig creates a new WorkerConfig with default values and applies the given options
func NewWorkerConfig(opts ...WorkerConfigOption) *WorkerConfig {
	config := &WorkerConfig{
		QueueSize:    100,
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		Enabled:      false,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// WithQueueSize sets the maximum queue size for the worker
func WithQueueSize(size int) WorkerConfigOption {
	return func(c *WorkerConfig) {
		c.QueueSize = size
	}
}

// WithBatchSize sets the maximum batch size for the worker
func WithBatchConfig(size int, timeout time.Duration) WorkerConfigOption {
	return func(c *WorkerConfig) {
		c.BatchSize = size
		c.Enabled = true
		c.BatchTimeout = timeout
	}
}
