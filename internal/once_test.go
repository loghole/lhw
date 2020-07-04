package internal

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOnce_Do(t *testing.T) {
	var (
		once = Once{}
		wg   = &sync.WaitGroup{}
	)

	var count int

	for i := 0; i < 5; i++ {
		wg.Add(1)

		go once.DoWG(wg, func() {
			count++
			time.Sleep(time.Second)
		})
	}

	wg.Wait()

	assert.Equal(t, 1, count, "expected 1")
}
