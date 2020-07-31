package internal

import (
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	queue := NewQueue(1)

	err := queue.Push([]byte("1"))
	assert.Nil(t, err, "expected nil error")

	err = queue.Push([]byte("2"))
	assert.EqualError(t, err, ErrQueueIsFull.Error())

	queue.Close()

	err = queue.Push([]byte("3"))
	assert.EqualError(t, err, ErrQueueIsClosed.Error())

	assert.Equal(t, []byte("1"), <-queue.Read())

	assert.Nil(t, <-queue.Read(), "expected nil")

	queue = NewQueue(10)
	err = queue.Push([]byte("1"))
	assert.Nil(t, err, "expected nil error")

	err = queue.Push([]byte("2"))
	assert.Nil(t, err, "expected nil error")

	err = queue.Push([]byte("3"))
	assert.Nil(t, err, "expected nil error")

	err = queue.Push([]byte("4"))
	assert.Nil(t, err, "expected nil error")

	queue.Close()

	assert.Equal(t, []byte("1"), <-queue.Read())

	assert.Equal(t, []byte("2"), <-queue.Read())

	assert.Equal(t, []byte("3"), <-queue.Read())

	assert.Equal(t, []byte("4"), <-queue.Read())
}

func TestQueueRace(t *testing.T) {
	queue := NewQueue(1000)

	var (
		wg      = &sync.WaitGroup{}
		counter int64
	)

	wg.Add(1)
	go func(counter *int64, wg *sync.WaitGroup, queue *Queue) {
		defer wg.Done()

		for range <-queue.Read() {
			atomic.AddInt64(counter, -1)
		}
	}(&counter, wg, queue)

	wg.Add(1)
	go func(counter *int64, wg *sync.WaitGroup, queue *Queue) {
		defer wg.Done()

		for {
			if err := queue.Push([]byte("msg")); err != nil {
				t.Log(err)
				return
			}
			atomic.AddInt64(counter, 1)
		}
	}(&counter, wg, queue)

	wg.Add(1)
	go func(counter *int64, wg *sync.WaitGroup, queue *Queue) {
		defer wg.Done()

		for {
			if err := queue.Push([]byte("msg")); err != nil {
				t.Log(err)
				return
			}
			atomic.AddInt64(counter, 1)
		}
	}(&counter, wg, queue)

	queue.Close()

	time.Sleep(time.Second)

	wg.Wait()

	assert.Equal(t, (int64)(0), counter)

	log.Println(counter)
}
