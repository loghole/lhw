package transport

import (
	"sync/atomic"
	"time"

	"github.com/loghole/lhw/internal"
)

type Transport interface {
	Send(body []byte) error
	IsConnected() bool
	IsReconnected() <-chan struct{}
}

type Config struct {
	Servers        []string
	Insecure       bool
	RequestTimeout time.Duration
	PingInterval   time.Duration
	SuccessCodes   []int
}

type httpTransport struct {
	connStatus  int32
	clientsPool ClientsPool

	requestTimeout time.Duration
	pingInterval   time.Duration
	successCodes   map[int]bool

	deadSignal internal.Signal
	liveSignal internal.Signal
}

func New(config Config) (Transport, error) {
	pool, err := NewClientsPool(config.Servers, config.Insecure)
	if err != nil {
		return nil, err
	}

	transport := &httpTransport{
		clientsPool:    pool,
		connStatus:     isLive,
		pingInterval:   config.PingInterval,
		requestTimeout: config.RequestTimeout,
		successCodes:   make(map[int]bool),

		liveSignal: make(internal.Signal, 1),
		deadSignal: make(internal.Signal, 1),
	}

	for _, code := range config.SuccessCodes {
		transport.successCodes[code] = true
	}

	go transport.pingDeadNodes()

	return transport, nil
}

func (t *httpTransport) IsConnected() (ok bool) {
	return atomic.LoadInt32(&t.connStatus) == isLive
}

func (t *httpTransport) IsReconnected() <-chan struct{} {
	return t.liveSignal
}

func (t *httpTransport) Send(body []byte) error {
	var (
		client *NodeClient
		code   int
		err    error
	)

	for {
		client, err = t.clientsPool.NextLive()
		if err != nil {
			atomic.StoreInt32(&t.connStatus, isDead)

			t.deadSignal.Send()

			return err
		}

		code, err = client.SendRequest(body, t.requestTimeout)
		if err == nil && t.successCodes[code] {
			return nil
		}

		t.clientsPool.OnFailure(client)
		t.deadSignal.Send()
	}
}

func (t *httpTransport) pingDeadNodes() {
	var (
		client *NodeClient
		code   int
		err    error
	)

	for {
		client, err = t.clientsPool.NextDead()
		if err != nil {
			<-t.deadSignal
			continue
		}

		code, err = client.PingRequest(t.requestTimeout)
		if err == nil && t.successCodes[code] {
			t.clientsPool.OnSuccess(client)

			atomic.StoreInt32(&t.connStatus, isLive)

			t.liveSignal.Send()
		}

		time.Sleep(t.pingInterval)
	}
}
