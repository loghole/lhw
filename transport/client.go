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
	storeURI = "/api/v1/store"
	pingURI  = "/api/v1/ping"

	authorizationHeader = "Authorization"
)

type NodeConfig struct {
	Host      string
	AuthToken string
}

type NodeClient struct {
	host  string
	token string

	status      int32
	activeReq   int32
	lastUseTime int64

	client *http.Client
}

// NewNodeClient create log hole node client.
func NewNodeClient(config NodeConfig, transport *http.Transport) *NodeClient {
	client := &NodeClient{
		host:   config.Host,
		token:  strings.Join([]string{"Bearer", config.AuthToken}, " "),
		status: isLive,
		client: &http.Client{Transport: transport},
	}

	return client
}

func (c *NodeClient) SendRequest(body []byte, timeout time.Duration) (code int, err error) {
	return c.do(storeURI, body, timeout)
}

// Ping request allows to check connection status.
func (c *NodeClient) PingRequest(timeout time.Duration) (code int, err error) {
	return c.do(pingURI, nil, timeout)
}

// ActiveRequests returns all active request of node client.
func (c *NodeClient) ActiveRequests() int {
	return int(atomic.LoadInt32(&c.activeReq))
}

// LastUseTime returns time of last started request.
func (c *NodeClient) LastUseTime() int {
	return int(atomic.LoadInt64(&c.lastUseTime))
}

func (c *NodeClient) do(uri string, body []byte, timeout time.Duration) (code int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	atomic.AddInt32(&c.activeReq, 1)
	defer atomic.AddInt32(&c.activeReq, -1)

	url := strings.Join([]string{c.host, uri}, "")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}

	req.Header.Set(authorizationHeader, c.token)

	atomic.StoreInt64(&c.lastUseTime, time.Now().UnixNano())

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}

	if err := resp.Body.Close(); err != nil {
		return 0, err
	}

	return resp.StatusCode, err
}
