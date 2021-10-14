package batchqueue

import (
	"sync"
	"sync/atomic"
)

const (
	// MaxMessageSize limit message size for transfer.
	MaxMessageSize = 5 * 1024 * 1024
	// MaxBatchSize will be the largest size for a batch sent from this particular producer.
	// This is used as a baseline to allocate a new buffer that can hold the entire batch
	// without needing costly re-allocations.
	MaxBatchSize = 128 * 1024
	// DefaultMaxMessagesPerBatch init default num of entries in per batch.
	DefaultMaxMessagesPerBatch = 1000
)

var (
	DefaultSequenceID = uint64(0)
)

// BatchBuilder wraps the objects needed to build a batch.
type BatchBuilder struct {
	buffer []interface{}

	// Current number of messages in the batch.
	numMessages uint

	// Max number of message allowed in the batch.
	maxMessages uint

	sequenceIDGenerator *uint64

	lock       sync.Mutex
	sequenceID uint64
}

// NewBatchBuilder init batch builder and return BatchBuilder pointer. Build a new batch message container.
func NewBatchBuilder(maxMessages uint) *BatchBuilder {
	if maxMessages == 0 {
		maxMessages = DefaultMaxMessagesPerBatch
	}

	bb := &BatchBuilder{
		buffer:              make([]interface{}, 0, maxMessages),
		numMessages:         0,
		maxMessages:         maxMessages,
		sequenceIDGenerator: &DefaultSequenceID,
	}
	return bb
}

// IsFull check if the size in the current batch exceeds the maximum size allowed by the batch.
func (bb *BatchBuilder) IsFull() bool {
	return bb.numMessages >= bb.maxMessages
}

// Add will add single message to batch.
func (bb *BatchBuilder) Add(payload interface{}) (isFull bool) {
	bb.lock.Lock()
	defer bb.lock.Unlock()
	bb.buffer = append(bb.buffer, payload)
	bb.numMessages++
	return bb.IsFull()
}

func (bb *BatchBuilder) reset() {
	bb.numMessages = 0
	bb.buffer = nil
}

// Flush all the messages buffered in the client and wait until all messages have been successfully persisted.
func (bb *BatchBuilder) Flush() (batchData []interface{}, sequenceID uint64) {
	if bb.numMessages == 0 {
		// No-Op for empty batch
		return nil, bb.sequenceID
	}
	bb.lock.Lock()
	defer bb.lock.Unlock()
	bb.sequenceID = GetAndAdd(bb.sequenceIDGenerator, 1)
	slice := bb.buffer
	bb.reset()

	return slice, bb.sequenceID
}

// GetAndAdd perform atomic read and update.
func GetAndAdd(n *uint64, diff uint64) uint64 {
	for {
		v := *n
		if atomic.CompareAndSwapUint64(n, v, v+diff) {
			return v
		}
	}
}
