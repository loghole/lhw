package transport

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func TestNodeClient_BulkRequest(t *testing.T) {
	const (
		host      = "127.0.0.1:8080"
		useragent = "test-client"
	)

	tests := []struct {
		name         string
		handler      func(t *testing.T) fasthttp.RequestHandler
		body         []byte
		timeout      time.Duration
		wantErr      bool
		expectedCode int
		expectedErr  error
	}{
		{
			name: "pass",
			handler: func(t *testing.T) fasthttp.RequestHandler {
				return func(ctx *fasthttp.RequestCtx) {
					assert.Equal(t, string(ctx.Request.Host()), host)
					assert.Equal(t, string(ctx.Method()), "POST")
					assert.Equal(t, string(ctx.Path()), "/api/v1/store/list")
					assert.Equal(t, string(ctx.UserAgent()), useragent)
					assert.Equal(t, string(ctx.Request.Header.ContentType()), "application/json")
					assert.Equal(t, string(ctx.Request.Body()), "BulkRequest")

					ctx.Response.Header.SetStatusCode(200)
				}
			},
			body:         []byte("BulkRequest"),
			timeout:      time.Second,
			wantErr:      false,
			expectedCode: 200,
		},
		{
			name: "timeout",
			handler: func(t *testing.T) fasthttp.RequestHandler {
				return func(ctx *fasthttp.RequestCtx) {
					assert.Equal(t, string(ctx.Request.Host()), host)
					assert.Equal(t, string(ctx.Method()), "POST")
					assert.Equal(t, string(ctx.Path()), "/api/v1/store/list")
					assert.Equal(t, string(ctx.UserAgent()), useragent)
					assert.Equal(t, string(ctx.Request.Header.ContentType()), "application/json")
					assert.Equal(t, string(ctx.Request.Body()), "BulkRequest")

					time.Sleep(2 * time.Second)
				}
			},
			body:         []byte("BulkRequest"),
			timeout:      time.Second,
			wantErr:      true,
			expectedCode: 200,
			expectedErr:  fasthttp.ErrTimeout,
		},
		{
			name: "connection closed",
			handler: func(t *testing.T) fasthttp.RequestHandler {
				return func(ctx *fasthttp.RequestCtx) {
					assert.Equal(t, string(ctx.Request.Host()), host)
					assert.Equal(t, string(ctx.Method()), "POST")
					assert.Equal(t, string(ctx.Path()), "/api/v1/store/list")
					assert.Equal(t, string(ctx.UserAgent()), useragent)
					assert.Equal(t, string(ctx.Request.Header.ContentType()), "application/json")
					assert.Equal(t, string(ctx.Request.Body()), "BulkRequest")

					ctx.Response.Header.SetStatusCode(500)
				}
			},
			body:         []byte("BulkRequest"),
			timeout:      time.Second,
			wantErr:      false,
			expectedCode: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				listener = fasthttputil.NewInmemoryListener()
				server   = fasthttp.Server{}
			)

			server.Handler = tt.handler(t)

			go server.Serve(listener)

			client := NewNodeClient(host, useragent)
			client.client.Dial = func(addr string) (conn net.Conn, err error) { return listener.Dial() }

			code, err := client.BulkRequest(tt.body, tt.timeout)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			assert.Equal(t, tt.expectedCode, code)

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			}

			listener.Close()
			server.Shutdown()
		})
	}
}

func TestNodeClient_PingRequest(t *testing.T) {
	const (
		host      = "127.0.0.1:8080"
		useragent = "test-client"
	)

	tests := []struct {
		name         string
		handler      func(t *testing.T) fasthttp.RequestHandler
		timeout      time.Duration
		wantErr      bool
		expectedCode int
		expectedErr  error
	}{
		{
			name: "pass",
			handler: func(t *testing.T) fasthttp.RequestHandler {
				return func(ctx *fasthttp.RequestCtx) {
					assert.Equal(t, string(ctx.Request.Host()), host)
					assert.Equal(t, string(ctx.Method()), "POST")
					assert.Equal(t, string(ctx.Path()), "/api/v1/ping")
					assert.Equal(t, string(ctx.UserAgent()), useragent)
					assert.Equal(t, string(ctx.Request.Header.ContentType()), "application/x-www-form-urlencoded")
					assert.Equal(t, string(ctx.Request.Body()), "")

					ctx.Response.Header.SetStatusCode(200)
				}
			},
			timeout:      time.Second,
			wantErr:      false,
			expectedCode: 200,
			expectedErr:  nil,
		},
		{
			name: "timeout",
			handler: func(t *testing.T) fasthttp.RequestHandler {
				return func(ctx *fasthttp.RequestCtx) {
					assert.Equal(t, string(ctx.Request.Host()), host)
					assert.Equal(t, string(ctx.Method()), "POST")
					assert.Equal(t, string(ctx.Path()), "/api/v1/ping")
					assert.Equal(t, string(ctx.UserAgent()), useragent)
					assert.Equal(t, string(ctx.Request.Header.ContentType()), "application/x-www-form-urlencoded")
					assert.Equal(t, string(ctx.Request.Body()), "")

					time.Sleep(2 * time.Second)
				}
			},
			timeout:      time.Second,
			wantErr:      true,
			expectedCode: 200,
			expectedErr:  fasthttp.ErrTimeout,
		},
		{
			name: "code 500",
			handler: func(t *testing.T) fasthttp.RequestHandler {
				return func(ctx *fasthttp.RequestCtx) {
					assert.Equal(t, string(ctx.Request.Host()), host)
					assert.Equal(t, string(ctx.Method()), "POST")
					assert.Equal(t, string(ctx.Path()), "/api/v1/ping")
					assert.Equal(t, string(ctx.UserAgent()), useragent)
					assert.Equal(t, string(ctx.Request.Header.ContentType()), "application/x-www-form-urlencoded")
					assert.Equal(t, string(ctx.Request.Body()), "")

					ctx.Response.Header.SetStatusCode(500)
				}
			},
			timeout:      time.Second,
			wantErr:      false,
			expectedCode: 500,
			expectedErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				listener = fasthttputil.NewInmemoryListener()
				server   = fasthttp.Server{}
			)

			server.Handler = tt.handler(t)

			go server.Serve(listener)

			client := NewNodeClient(host, useragent)
			client.client.Dial = func(addr string) (conn net.Conn, err error) { return listener.Dial() }

			code, err := client.PingRequest(tt.timeout)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			assert.Equal(t, tt.expectedCode, code)

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr.Error())
			}

			listener.Close()
			server.Shutdown()
		})
	}
}

func TestNodeClient_LastUseTime(t *testing.T) {
	const (
		host      = "127.0.0.1:8080"
		useragent = "test-client"
	)

	var (
		listener = fasthttputil.NewInmemoryListener()
		server   = fasthttp.Server{}
	)

	server.Handler = func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetStatusCode(200)
	}

	go server.Serve(listener)

	client := NewNodeClient(host, useragent)
	client.client.Dial = func(addr string) (conn net.Conn, err error) { return listener.Dial() }

	startTime := client.LastUseTime()

	_, err := client.PingRequest(time.Second)
	if err != nil {
		t.Error(err)
	}

	assert.NotEqual(t, client.LastUseTime(), startTime, "start time as equals time after request")

	listener.Close()
	server.Shutdown()
}

func TestNodeClient_PendingRequests(t *testing.T) {
	const (
		host      = "http://127.0.0.1:8080"
		useragent = "test-client"
	)

	var (
		listener = fasthttputil.NewInmemoryListener()
		server   = fasthttp.Server{}
	)

	server.Handler = func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetStatusCode(200)
		time.Sleep(5 * time.Second)
	}

	go server.Serve(listener)

	client := NewNodeClient(host, useragent)
	client.client.Dial = func(addr string) (conn net.Conn, err error) { return listener.Dial() }

	assert.Equal(t, 0, client.PendingRequests(), "expected 0 pending requests")

	for i := 0; i < 5; i++ {
		client.PingRequest(100 * time.Millisecond)
	}

	assert.Equal(t, 5, client.PendingRequests(), "expected 5 pending requests")

	listener.Close()
	server.Shutdown()
}
