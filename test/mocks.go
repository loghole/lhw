package test

import (
	"fmt"

	"github.com/stretchr/testify/mock"

	"github.com/gadavy/lhw/internal"
)

type MockTransport struct {
	mock.Mock
	isConnectedCounter int
}

func (m *MockTransport) Send(body []byte) error {
	return m.Called(body).Error(0)
}

func (m *MockTransport) IsConnected() (ok bool) {
	ok = m.Called().Bool(0 + m.isConnectedCounter)
	m.isConnectedCounter++

	return ok
}

func (m *MockTransport) IsReconnected() <-chan struct{} {
	return m.Called().Get(0).(<-chan struct{})
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Printf(format string, v ...interface{}) {
	m.Called(fmt.Sprintf(format, v...))
}

type StubTransport struct {
	Ch internal.Signal
}

func (m *StubTransport) SendBulk([]byte) error          { return nil }
func (m *StubTransport) IsConnected() bool              { return true }
func (m *StubTransport) IsReconnected() <-chan struct{} { return m.Ch }

type StubStorage struct{}

func (m *StubStorage) Put([]byte) error     { return nil }
func (m *StubStorage) Pop() ([]byte, error) { return nil, nil }
func (m *StubStorage) Drop() error          { return nil }
func (m *StubStorage) IsUsed() bool         { return false }
