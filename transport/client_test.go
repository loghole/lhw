package transport

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNodeClient_ActiveRequests(t *testing.T) {
	tests := []struct {
		name        string
		input       *NodeClient
		expectedRes int
	}{
		{
			name:        "#1",
			input:       &NodeClient{activeReq: 1},
			expectedRes: 1,
		},
		{
			name:        "#2",
			input:       &NodeClient{activeReq: 133},
			expectedRes: 133,
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expectedRes, tt.input.ActiveRequests())
	}
}

func TestNodeClient_LastUseTime(t *testing.T) {
	tests := []struct {
		name        string
		input       *NodeClient
		expectedRes int
	}{
		{
			name:        "#1",
			input:       &NodeClient{lastUseTime: 1594109106986231572},
			expectedRes: 1594109106986231572,
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expectedRes, tt.input.LastUseTime())
	}
}

func TestNodeClient_PingRequest(t *testing.T) {
	ts := httptest.NewUnstartedServer(nil)
	ts.EnableHTTP2 = true
	ts.StartTLS()
	defer ts.Close()

	tests := []struct {
		name         string
		handler      http.HandlerFunc
		wantErr      bool
		expectedCode int
		expectedErr  string
	}{
		{
			name:  "Code200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, pingURI, r.URL.String())
				assert.Equal(t, http.MethodPost, r.Method)

				w.WriteHeader(http.StatusOK)
			},
			expectedCode: 200,
		},
		{
			name:  "Code500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, pingURI, r.URL.String())
				assert.Equal(t, http.MethodPost, r.Method)

				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedCode: 500,
		},
		{
			name:  "Timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Second * 2)
			},
			wantErr:      true,
			expectedCode: 0,
			expectedErr:  fmt.Sprintf("Post \"%s/api/v1/ping\": context deadline exceeded", ts.URL),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = tt.handler

			client := NodeClient{addr: ts.URL, client: ts.Client()}

			code, err := client.PingRequest(time.Second)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr)
			}

			assert.Equal(t, tt.expectedCode, code)
		})
	}
}

func TestNodeClient_SendRequest(t *testing.T) {
	ts := httptest.NewUnstartedServer(nil)
	ts.EnableHTTP2 = true
	ts.StartTLS()
	defer ts.Close()

	tests := []struct {
		name         string
		body         []byte
		handler      http.HandlerFunc
		wantErr      bool
		expectedCode int
		expectedErr  string
	}{
		{
			name:  "Code200",
			body: []byte(`{"message":"some message"}`),
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, storeURI, r.URL.String())
				assert.Equal(t, http.MethodPost, r.Method)

				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}

				assert.Equal(t, []byte(`{"message":"some message"}`), body)

				w.WriteHeader(http.StatusOK)
			},
			expectedCode: 200,
		},
		{
			name:  "Code500",
			body: []byte(`{"message":"some message"}`),
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, storeURI, r.URL.String())
				assert.Equal(t, http.MethodPost, r.Method)

				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}

				assert.Equal(t, []byte(`{"message":"some message"}`), body)

				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedCode: 500,
		},
		{
			name:  "Timeout",
			body: []byte(`{"message":"some message"}`),
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, storeURI, r.URL.String())
				assert.Equal(t, http.MethodPost, r.Method)

				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}

				assert.Equal(t, []byte(`{"message":"some message"}`), body)

				time.Sleep(time.Second * 2)
			},
			wantErr:      true,
			expectedCode: 0,
			expectedErr:  fmt.Sprintf("Post \"%s/api/v1/store\": context deadline exceeded", ts.URL),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = tt.handler

			client := NodeClient{addr: ts.URL, client: ts.Client()}

			code, err := client.SendRequest(tt.body, time.Second)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr)
			}

			assert.Equal(t, tt.expectedCode, code)
		})
	}
}