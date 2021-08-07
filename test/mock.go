package test

import (
	"sync/atomic"
)

type StubTransport struct {
	Counter int64
}

func (m *StubTransport) Send(body []byte) error {
	atomic.AddInt64(&m.Counter, 1)

	return nil
}

func (m *StubTransport) IsConnected() (ok bool) {
	return true
}

func (m *StubTransport) IsReconnected() <-chan struct{} {
	return nil
}
