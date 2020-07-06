package transport

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

const (
	isDead int32 = iota
	isLive
)

const (
	storeBatchURI = "/api/v1/store"
	pingServerURI = "/api/v1/ping"
)

type NodeClient struct {
	host string

	status      int32
	pendingReq  int32
	lastUseTime int64

	client http.Client
}

// NewNodeClient create elastic node client with small api.
func NewNodeClient(url string, transport *http.Transport) *NodeClient {
	client := &NodeClient{
		host:   url,
		status: isLive,
		client: http.Client{Transport: transport},
	}

	return client
}

func (c *NodeClient) SendRequest(body []byte, timeout time.Duration) (code int, err error) {
	return c.do(storeBatchURI, body, timeout)
}

// Ping request allows to check connection status.
func (c *NodeClient) PingRequest(timeout time.Duration) (code int, err error) {
	return c.do(pingServerURI, nil, timeout)
}

// ActiveRequests returns all active request of node client.
func (c *NodeClient) ActiveRequests() int {
	return int(atomic.LoadInt32(&c.pendingReq))
}

// LastUseTime returns time of last started request.
func (c *NodeClient) LastUseTime() int {
	return int(atomic.LoadInt64(&c.lastUseTime))
}

func (c *NodeClient) do(uri string, body []byte, timeout time.Duration) (code int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	atomic.AddInt32(&c.pendingReq, 1)
	defer atomic.AddInt32(&c.pendingReq, -1)

	url := strings.Join([]string{c.host, uri}, "")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}

	atomic.StoreInt64(&c.lastUseTime, time.Now().UnixNano())

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, err
}
