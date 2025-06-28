package shared

import (
	"fmt"
)

type JobFunc func()

type JobQueue interface {
	Enqueue(job JobFunc)
}

// SerialJobQueue is a simple serial job queue implementation.
type SerialJobQueue struct {
	queue chan JobFunc
}

func NewSerialJobQueue(buffer int) *SerialJobQueue {
	jq := &SerialJobQueue{
		queue: make(chan JobFunc, buffer),
	}
	go jq.worker()
	return jq
}

func (jq *SerialJobQueue) worker() {
	for job := range jq.queue {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("job panic: %v\n", r)
				}
			}()
			job()
		}()
	}
}

func (jq *SerialJobQueue) Enqueue(job JobFunc) {
	jq.queue <- job
}

var _ JobQueue = (*SerialJobQueue)(nil)
