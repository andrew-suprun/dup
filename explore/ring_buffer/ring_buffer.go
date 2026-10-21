package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrBufferFull  = errors.New("ring buffer is full")
	ErrBufferEmpty = errors.New("ring buffer is empty")
)

// Consumer represents an individual consumer tracking its own read position.
type Consumer struct {
	cursor uint64
}

// RingBuffer is a generic SPMC ring buffer.
type RingBuffer[T any] struct {
	buf      []T
	capacity uint64
	head     uint64 // Written to only by the single producer
}

// NewRingBuffer creates a ring buffer with the given capacity (must be a power of 2 for optimal performance).
func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	// Ensure capacity is a power of 2 for fast bitwise modulo operations
	capPow2 := uint64(1)
	for capPow2 < uint64(capacity) {
		capPow2 <<= 1
	}

	return &RingBuffer[T]{
		buf:      make([]T, capPow2),
		capacity: capPow2,
	}
}

// NewConsumer registers a new consumer for the ring buffer.
// It starts reading from the current buffer head (or you can adjust its initial cursor).
func (rb *RingBuffer[T]) NewConsumer() *Consumer {
	return &Consumer{
		cursor: atomic.LoadUint64(&rb.head),
	}
}

// Push adds an item to the buffer. Only safe for a SINGLE producer.
// Returns ErrBufferFull if the slowest consumer hasn't kept up and would be overwritten.
func (rb *RingBuffer[T]) Push(item T, consumers []*Consumer) error {
	head := rb.head

	// Check if any consumer is too far behind (buffer is full relative to that consumer)
	if len(consumers) > 0 {
		minCursor := head
		for _, c := range consumers {
			cPos := atomic.LoadUint64(&c.cursor)
			if cPos < minCursor {
				minCursor = cPos
			}
		}
		// If the distance between the slowest consumer and head equals capacity, we are full
		if head-minCursor >= rb.capacity {
			return ErrBufferFull
		}
	}

	// Write item at the head position (using bitwise AND for ring wrapping)
	rb.buf[head&(rb.capacity-1)] = item

	// Advance the head pointer
	atomic.StoreUint64(&rb.head, head+1)
	return nil
}

// Pop retrieves the next item for a specific consumer.
// Returns ErrBufferEmpty if the consumer has caught up to the producer.
func (rb *RingBuffer[T]) Pop(c *Consumer) (T, error) {
	var zero T
	head := atomic.LoadUint64(&rb.head)
	cursor := atomic.LoadUint64(&c.cursor)

	if cursor >= head {
		return zero, ErrBufferEmpty
	}

	// Read item at the consumer's cursor position
	item := rb.buf[cursor&(rb.capacity-1)]

	// Advance consumer's cursor
	atomic.StoreUint64(&c.cursor, cursor+1)

	return item, nil
}

func main() {
	// Create a ring buffer with capacity 8
	rb := NewRingBuffer[int](8)

	// Register two consumers
	c1 := rb.NewConsumer()
	c2 := rb.NewConsumer()
	consumers := []*Consumer{c1, c2}

	var wg sync.WaitGroup
	wg.Add(3) // 1 producer + 2 consumers

	// Producer Goroutine
	go func() {
		defer wg.Done()
		for i := 1; i <= 100; i++ {
			for {
				err := rb.Push(i, consumers)
				if err == nil {
					fmt.Printf("[Producer] Pushed: %d\n", i)
					break
				}
				// If full, yield briefly and retry
				time.Sleep(2 * time.Millisecond)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Consumer 1 Goroutine
	go func() {
		defer wg.Done()
		for {
			val, err := rb.Pop(c1)
			if err == nil {
				fmt.Printf("[Consumer 1] Read: %d\n", val)
				if val == 100 {
					break
				}
			} else {
				time.Sleep(15 * time.Millisecond)
			}
		}
	}()

	// Consumer 2 Goroutine
	go func() {
		defer wg.Done()
		for {
			val, err := rb.Pop(c2)
			if err == nil {
				fmt.Printf("[Consumer 2] Read: %d\n", val)
				if val == 100 {
					break
				}
			} else {
				time.Sleep(18 * time.Millisecond)
			}
		}
	}()

	wg.Wait()
	fmt.Println("Done!")
}
