package internal

import (
	"errors"
	"sync"
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
	ch    chan []byte
	cond  *sync.Cond
	state int32
}

func NewQueue(capacity int) *Queue {
	return &Queue{
		ch:    make(chan []byte, capacity),
		cond:  sync.NewCond(&sync.Mutex{}),
		state: opened,
	}
}

func (s *Queue) Push(data []byte) error {
	if atomic.LoadInt32(&s.state) != opened {
		return ErrQueueIsClosed
	}

	select {
	case s.ch <- data:
		s.cond.Signal()
		return nil
	default:
		return ErrQueueIsFull
	}
}

func (s *Queue) Read() []byte {
	return <-s.ch
}

func (s *Queue) Close() {
	atomic.StoreInt32(&s.state, closed)
	s.cond.Signal()
}

func (s *Queue) Next() bool {
	s.cond.L.Lock()

	if atomic.LoadInt32(&s.state) == opened && len(s.ch) == 0 {
		s.cond.Wait()
	}

	s.cond.L.Unlock()

	return atomic.LoadInt32(&s.state) == opened || len(s.ch) > 0
}
