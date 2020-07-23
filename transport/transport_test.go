package transport

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
				Nodes: nil,
			},
			wantErr:     true,
			expectedErr: "no servers available for connection",
		},
		{
			name: "Pass",
			cfg: Config{
				Nodes:          []NodeConfig{{Host: "http://127.0.0.1:9200"}},
				RequestTimeout: time.Hour,
				PingInterval:   time.Hour,
				SuccessCodes:   []int{200, 201, 202},
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
	tests := []struct {
		name        string
		input       []byte
		transport   *httpTransport
		client      *NodeClient
		handler     http.HandlerFunc
		wantErr     bool
		expectedErr string
	}{
		{
			name:  "PoolError",
			input: []byte(`{"message":"some message"}"`),
			transport: &httpTransport{
				requestTimeout: time.Second,
				pingInterval:   time.Second,
				successCodes:   map[int]bool{200: true},
				deadSignal:     make(internal.Signal, 1),
				liveSignal:     make(internal.Signal, 1),
			},
			client: &NodeClient{
				status: isDead,
			},
			handler:     func(w http.ResponseWriter, r *http.Request) {},
			wantErr:     true,
			expectedErr: ErrNoAvailableClients.Error(),
		},
		{
			name:  "RequestError",
			input: []byte(`{"message":"some message"}"`),
			transport: &httpTransport{
				requestTimeout: time.Second,
				pingInterval:   time.Second,
				successCodes:   map[int]bool{200: true},
				deadSignal:     make(internal.Signal, 1),
				liveSignal:     make(internal.Signal, 1),
			},
			client: &NodeClient{
				status: isLive,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, storeURI, r.URL.String())
				assert.Equal(t, `{"message":"some message"}"`, bodyString(r))

				time.Sleep(5 * time.Second)
			},
			wantErr:     true,
			expectedErr: ErrNoAvailableClients.Error(),
		},
		{
			name:  "RequestPass",
			input: []byte(`{"message":"some message"}"`),
			transport: &httpTransport{
				requestTimeout: time.Second,
				pingInterval:   time.Second,
				successCodes:   map[int]bool{200: true},
				deadSignal:     make(internal.Signal, 1),
				liveSignal:     make(internal.Signal, 1),
			},
			client: &NodeClient{
				status: isLive,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, storeURI, r.URL.String())
				assert.Equal(t, `{"message":"some message"}"`, bodyString(r))

				w.WriteHeader(http.StatusOK)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewUnstartedServer(tt.handler)
			ts.EnableHTTP2 = true
			ts.StartTLS()

			tt.client.client = ts.Client()
			tt.client.host = ts.URL

			transport := tt.transport
			transport.clientsPool = &SinglePool{client: tt.client}

			err := transport.Send(tt.input)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr)
			}

			ts.Close()
		})
	}
}

func TestHttpTransport_pingDeadNodes(t *testing.T) {
	tests := []struct {
		name        string
		transport   *httpTransport
		client      *NodeClient
		handler     http.HandlerFunc
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
				status: isLive,
			},
			handler:     func(w http.ResponseWriter, r *http.Request) {},
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
				status: isDead,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, pingURI, r.URL.String())

				w.WriteHeader(http.StatusOK)
			},
			reconnected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewUnstartedServer(tt.handler)
			ts.EnableHTTP2 = true
			ts.StartTLS()

			tt.client.client = ts.Client()
			tt.client.host = ts.URL

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

			ts.Close()
		})
	}
}

func bodyString(r *http.Request) string {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	return string(data)
}
