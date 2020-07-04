package transport

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	isDead uint32 = iota
	isLive
)

const (
	contentType   = "application/json"
	storeBatchURI = "/api/v1/store/list"
	pingServerURI = "/api/v1/ping"
)

const (
	MaxIdleConnDuration = 5 * time.Second
)

type NodeClient struct {
	host      string
	useragent string

	status      uint32
	lastUseTime int64

	client fasthttp.HostClient
}

// NewNodeClient create elastic node client with small api.
func NewNodeClient(url, useragent string) *NodeClient {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")

	client := &NodeClient{
		host:      url,
		useragent: useragent,
		status:    isLive,
		client: fasthttp.HostClient{
			Addr:                url,
			MaxIdleConnDuration: MaxIdleConnDuration,
		},
	}

	return client
}

func (c *NodeClient) BulkRequest(body []byte, timeout time.Duration) (code int, err error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetUserAgent(c.useragent)
	req.Header.SetContentType(contentType)
	req.Header.SetRequestURI(storeBatchURI)
	req.Header.SetHost(c.host)

	req.SetBody(body)

	atomic.StoreInt64(&c.lastUseTime, time.Now().UnixNano())

	err = c.client.DoTimeout(req, resp, timeout)

	return resp.StatusCode(), err
}

// Ping request allows to check connection status.
func (c *NodeClient) PingRequest(timeout time.Duration) (code int, err error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetUserAgent(c.useragent)
	req.Header.SetRequestURI(pingServerURI)
	req.Header.SetHost(c.host)

	atomic.StoreInt64(&c.lastUseTime, time.Now().UnixNano())

	err = c.client.DoTimeout(req, resp, timeout)

	return resp.StatusCode(), err
}

// PendingRequests returns all pending request of node client.
func (c *NodeClient) PendingRequests() int {
	return c.client.PendingRequests()
}

// LastUseTime returns time of last started request.
func (c *NodeClient) LastUseTime() int {
	return int(atomic.LoadInt64(&c.lastUseTime))
}
