package internal

import (
	"errors"
	"sync/atomic"
)

var (
	ErrQueueIsClosed = errors.New("queue is closed")
	ErrQueueIsFull   = errors.New("queue is full")
)

const (
	opened int32 = iota + 1
	closed
)

type Queue struct {
	ch     chan []byte
	closed int32
}

func NewQueue(capacity int) *Queue {
	return &Queue{
		ch:     make(chan []byte, capacity),
		closed: opened,
	}
}

func (s *Queue) Push(data []byte) error {
	if !atomic.CompareAndSwapInt32(&s.closed, opened, atomic.LoadInt32(&s.closed)) {
		return ErrQueueIsClosed
	}

	select {
	case s.ch <- data:
		return nil
	default:
		return ErrQueueIsFull
	}
}

func (s *Queue) Read() <-chan []byte {
	return s.ch
}

func (s *Queue) Close() {
	if atomic.CompareAndSwapInt32(&s.closed, opened, closed) {
		close(s.ch)
	}
}
