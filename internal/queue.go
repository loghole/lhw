package internal

import (
	"errors"
	"sync"
)

var (
	ErrIsClosed = errors.New("is closed")
	ErrIsFull   = errors.New("is full")
)

type Queue struct {
	ch     chan []byte
	mu     sync.RWMutex
	wg     sync.WaitGroup
	sn     sync.Once
	closed bool
}

func NewQueue(cap int) *Queue {
	return &Queue{
		ch: make(chan []byte, cap),
	}
}

func (q *Queue) Push(data []byte) error {
	q.mu.RLock()
	if q.closed {
		q.mu.RUnlock()
		return ErrIsClosed
	}
	q.wg.Add(1)
	q.mu.RUnlock()

	select {
	case q.ch <- data:
		q.wg.Done()
		return nil
	default:
		q.wg.Done()
		return ErrIsFull
	}
}

func (q *Queue) Read() <-chan []byte {
	return q.ch
}

func (q *Queue) Close() {
	q.sn.Do(q.close)
}

func (q *Queue) close() {
	q.mu.Lock()
	q.closed = true
	q.wg.Wait()
	close(q.ch)
	q.mu.Unlock()
}
