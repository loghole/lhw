package transport

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	"github.com/gadavy/lhw/internal"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		cfg         Config
		wantErr     bool
		expectedRes interface{}
		expectedErr string
	}{
		{
			name: "Error",
			cfg: Config{
				NodeURIs: nil,
			},
			wantErr:     true,
			expectedErr: "no servers available for connection",
		},
		{
			name: "Pass",
			cfg: Config{
				NodeURIs:       []string{"http://127.0.0.1:9200"},
				RequestTimeout: time.Hour,
				PingInterval:   time.Hour,
				SuccessCodes:   []int{200, 201, 202},
				UserAgent:      "test-useragent",
			},
			wantErr: false,
			expectedRes: &httpTransport{
				connStatus:     isLive,
				requestTimeout: time.Hour,
				pingInterval:   time.Hour,
				successCodes:   map[int]bool{200: true, 201: true, 202: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			assert.IsType(t, tt.expectedRes, res)

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.Equal(t, tt.expectedRes.(*httpTransport).connStatus, res.(*httpTransport).connStatus)
				assert.Equal(t, tt.expectedRes.(*httpTransport).requestTimeout, res.(*httpTransport).requestTimeout)
				assert.Equal(t, tt.expectedRes.(*httpTransport).pingInterval, res.(*httpTransport).pingInterval)
				assert.Equal(t, tt.expectedRes.(*httpTransport).successCodes, res.(*httpTransport).successCodes)
			}
		})
	}
}

func TestHttpTransport_IsConnected(t *testing.T) {
	tests := []struct {
		name        string
		transport   *httpTransport
		expectedRes bool
	}{
		{
			name:        "Connected",
			transport:   &httpTransport{connStatus: isLive},
			expectedRes: true,
		},
		{
			name:        "Disconnected",
			transport:   &httpTransport{connStatus: isDead},
			expectedRes: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedRes, tt.transport.IsConnected())
		})
	}
}

func TestHttpTransport_IsReconnected(t *testing.T) {
	ch := make(internal.Signal, 1)

	tests := []struct {
		name        string
		transport   *httpTransport
		expectedRes <-chan struct{}
	}{
		{
			name:        "Connected",
			transport:   &httpTransport{liveSignal: ch},
			expectedRes: ch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedRes, tt.transport.IsReconnected())
		})
	}
}

func TestHttpTransport_SendBulk(t *testing.T) {
	const (
		host      = "http://127.0.0.1:8080"
		useragent = "test-client"
	)

	tests := []struct {
		name        string
		input       []byte
		transport   *httpTransport
		client      *NodeClient
		handler     func(t *testing.T) fasthttp.RequestHandler
		wantErr     bool
		expectedErr string
	}{
		{
			name:  "PoolError",
			input: []byte("bulk"),
			transport: &httpTransport{
				requestTimeout: time.Second,
				pingInterval:   time.Second,
				successCodes:   map[int]bool{200: true},
				deadSignal:     make(internal.Signal, 1),
				liveSignal:     make(internal.Signal, 1),
			},
			client: &NodeClient{
				host:      host,
				useragent: useragent,
				status:    isDead,
				client: fasthttp.HostClient{
					Addr:     "127.0.0.1:8080",
					MaxConns: 1,
				},
			},
			handler: func(t *testing.T) fasthttp.RequestHandler {
				return func(ctx *fasthttp.RequestCtx) {
					assert.Equal(t, string(ctx.Request.Host()), host)
					assert.Equal(t, string(ctx.UserAgent()), useragent)
					assert.Equal(t, string(ctx.Request.Body()), "bulk")
				}
			},
			wantErr:     true,
			expectedErr: ErrNoAvailableClients.Error(),
		},
		{
			name:  "RequestError",
			input: []byte("bulk"),
			transport: &httpTransport{
				requestTimeout: time.Second,
				pingInterval:   time.Second,
				successCodes:   map[int]bool{200: true},
				deadSignal:     make(internal.Signal, 1),
				liveSignal:     make(internal.Signal, 1),
			},
			client: &NodeClient{
				host:      host,
				useragent: useragent,
				status:    isLive,
				client: fasthttp.HostClient{
					Addr:     "127.0.0.1:8080",
					MaxConns: 1,
				},
			},
			handler: func(t *testing.T) fasthttp.RequestHandler {
				return func(ctx *fasthttp.RequestCtx) {
					assert.Equal(t, string(ctx.Request.Host()), host)
					assert.Equal(t, string(ctx.UserAgent()), useragent)
					assert.Equal(t, string(ctx.Request.Body()), "bulk")

					time.Sleep(5 * time.Second)
				}
			},
			wantErr:     true,
			expectedErr: ErrNoAvailableClients.Error(),
		},
		{
			name:  "RequestPass",
			input: []byte("bulk"),
			transport: &httpTransport{
				requestTimeout: time.Second,
				pingInterval:   time.Second,
				successCodes:   map[int]bool{200: true},
				deadSignal:     make(internal.Signal, 1),
				liveSignal:     make(internal.Signal, 1),
			},
			client: &NodeClient{
				host:      host,
				useragent: useragent,
				status:    isLive,
				client: fasthttp.HostClient{
					Addr:     "127.0.0.1:8080",
					MaxConns: 1,
				},
			},
			handler: func(t *testing.T) fasthttp.RequestHandler {
				return func(ctx *fasthttp.RequestCtx) {
					assert.Equal(t, string(ctx.Request.Host()), host)
					assert.Equal(t, string(ctx.UserAgent()), useragent)
					assert.Equal(t, string(ctx.Request.Body()), "bulk")

					ctx.Response.SetStatusCode(200)
				}
			},
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

			// init new client and transport
			tt.client.client.Dial = func(addr string) (conn net.Conn, err error) { return listener.Dial() }

			transport := tt.transport
			transport.clientsPool = &SinglePool{client: tt.client}

			err := transport.SendBulk(tt.input)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr)
			}

			listener.Close()
			server.Shutdown()
		})
	}
}

func TestHttpTransport_pingDeadNodes(t *testing.T) {
	const (
		host      = "http://127.0.0.1:8080"
		useragent = "test-client"
	)

	tests := []struct {
		name        string
		transport   *httpTransport
		client      *NodeClient
		handler     func(t *testing.T) fasthttp.RequestHandler
		reconnected bool
	}{
		{
			name: "PoolError",
			transport: &httpTransport{
				requestTimeout: time.Second,
				pingInterval:   time.Second,
				successCodes:   map[int]bool{200: true},
				deadSignal:     make(internal.Signal, 1),
				liveSignal:     make(internal.Signal, 1),
			},
			client: &NodeClient{
				host:      host,
				useragent: useragent,
				status:    isLive,
				client: fasthttp.HostClient{
					Addr:     "127.0.0.1:8080",
					MaxConns: 1,
				},
			},
			handler: func(t *testing.T) fasthttp.RequestHandler {
				return func(ctx *fasthttp.RequestCtx) {}
			},
			reconnected: false,
		},
		{
			name: "Pass",
			transport: &httpTransport{
				requestTimeout: time.Second,
				pingInterval:   time.Second,
				successCodes:   map[int]bool{200: true},
				deadSignal:     make(internal.Signal, 1),
				liveSignal:     make(internal.Signal, 1),
			},
			client: &NodeClient{
				host:      host,
				useragent: useragent,
				status:    isDead,
				client: fasthttp.HostClient{
					Addr:     "127.0.0.1:8080",
					MaxConns: 1,
				},
			},
			handler: func(t *testing.T) fasthttp.RequestHandler {
				return func(ctx *fasthttp.RequestCtx) {
					assert.Equal(t, string(ctx.Request.Host()), host)
					assert.Equal(t, string(ctx.UserAgent()), useragent)

					ctx.SetStatusCode(200)
				}
			},
			reconnected: true,
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

			// init new client and transport
			tt.client.client.Dial = func(addr string) (conn net.Conn, err error) { return listener.Dial() }

			transport := tt.transport
			transport.clientsPool = &SinglePool{client: tt.client}

			go transport.pingDeadNodes()

			transport.deadSignal.Send()

			if tt.reconnected {
				select {
				case <-time.After(time.Second):
					t.Error("reconnection failed")
				case <-transport.IsReconnected():
					assert.True(t, transport.IsConnected(), "transport should be connected")
					assert.True(t, tt.client.status == isLive, "client should be connected")

				}
			}

			listener.Close()
			server.Shutdown()
		})
	}
}
