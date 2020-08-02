package internal

import (
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
	assert.EqualError(t, err, ErrIsFull.Error())

	queue.Close()

	err = queue.Push([]byte("3"))
	assert.EqualError(t, err, ErrIsClosed.Error())

	assert.Equal(t, []byte("1"), <-queue.Read())

	assert.Equal(t, []byte(nil), <-queue.Read())

	queue = NewQueue(10)

	err = queue.Push([]byte("1"))
	assert.Nil(t, err, "expected nil error")

	err = queue.Push([]byte("2"))
	assert.Nil(t, err, "expected nil error")

	err = queue.Push([]byte("3"))
	assert.Nil(t, err, "expected nil error")

	err = queue.Push([]byte("4"))
	assert.Nil(t, err, "expected nil error")

	assert.Equal(t, []byte("1"), <-queue.Read())

	queue.Close()

	assert.Equal(t, []byte("2"), <-queue.Read())

	assert.Equal(t, []byte("3"), <-queue.Read())

	assert.Equal(t, []byte("4"), <-queue.Read())

	assert.Equal(t, []byte(nil), <-queue.Read())
}

func TestQueueRace(t *testing.T) {
	var (
		queue = NewQueue(1000)
		wg    = &sync.WaitGroup{}

		counter int64
	)

	wg.Add(3)
	go writer(wg, queue, &counter)
	go writer(wg, queue, &counter)
	go reader(wg, queue, &counter)

	queue.Close()

	time.AfterFunc(time.Second*5, func() { t.Fail() })

	wg.Wait()
	assert.Equal(t, (int64)(0), counter)
}

func writer(wg *sync.WaitGroup, queue *Queue, counter *int64) {
	defer wg.Done()

	for {
		if err := queue.Push([]byte("msg")); err != nil {
			return
		}

		atomic.AddInt64(counter, 1)
	}
}

func reader(wg *sync.WaitGroup, queue *Queue, counter *int64) {
	defer wg.Done()

	for range queue.Read() {
		atomic.AddInt64(counter, -1)
	}
}

func BenchmarkNewQueue(b *testing.B) {
	b.Run("Queue", func(b *testing.B) {
		queue := NewQueue(b.N)

		for i := 0; i < b.N; i++ {
			queue.Push([]byte{})
		}
	})

	b.Run("Chan", func(b *testing.B) {
		queue := make(chan []byte, b.N)

		for i := 0; i < b.N; i++ {
			queue <- []byte{}
		}
	})

	b.Run("QueueParallel (10)", func(b *testing.B) {
		b.SetParallelism(10)

		queue := NewQueue(b.N * 10)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				queue.Push([]byte{})
			}
		})
	})

	b.Run("ChanParallel (10)", func(b *testing.B) {
		b.SetParallelism(10)

		queue := make(chan []byte, b.N*10)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				queue <- []byte{}
			}
		})
	})
}
