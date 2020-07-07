package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignal_Send(t *testing.T) {
	var signal Signal = make(chan struct{}, 1)

	signal.Send()
	signal.Send()
	signal.Send()

	assert.Equal(t, struct{}{}, <-signal)
}
