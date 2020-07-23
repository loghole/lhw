package lhw

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gadavy/lhw/transport"
)

func TestOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   Option
		expected *writerConfig
	}{
		{
			name:     "WithQueueCap",
			option:   WithQueueCap(10),
			expected: &writerConfig{QueueCap: 10},
		},
		{
			name:     "WithLogger",
			option:   WithLogger(log.New(os.Stdout, "", log.Ltime)),
			expected: &writerConfig{Logger: log.New(os.Stdout, "", log.Ltime)},
		},
		{
			name:   "Node",
			option: Node("127.0.0.1:50000"),
			expected: &writerConfig{
				NodeConfigs: []transport.NodeConfig{
					{
						Host: "127.0.0.1:50000",
					},
				},
			},
		},
		{
			name:   "NodeWithAuth",
			option: NodeWithAuth("127.0.0.1:50000", "token"),
			expected: &writerConfig{
				NodeConfigs: []transport.NodeConfig{
					{
						Host:      "127.0.0.1:50000",
						AuthToken: "token",
					},
				},
			},
		},
		{
			name:     "WithInsecure",
			option:   WithInsecure(),
			expected: &writerConfig{Insecure: true},
		},
		{
			name:     "WithRequestTimeout",
			option:   WithRequestTimeout(time.Second),
			expected: &writerConfig{RequestTimeout: time.Second},
		},
		{
			name:     "WithPingInterval",
			option:   WithPingInterval(time.Second),
			expected: &writerConfig{PingInterval: time.Second},
		},
		{
			name:     "WithSuccessCodes",
			option:   WithSuccessCodes(200, 201),
			expected: &writerConfig{SuccessCodes: []int{200, 201}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &writerConfig{}

			tt.option(config)

			assert.Equal(t, tt.expected, config)
		})
	}
}

func TestBuildWriterConfig(t *testing.T) {
	tests := []struct {
		name        string
		options     []Option
		wantErr     bool
		expectedRes *writerConfig
		expectedErr string
	}{
		{
			name: "Pass#1",
			options: []Option{
				WithQueueCap(10),
				WithLogger(log.New(os.Stdout, "", log.Ltime)),
				Node("127.0.0.1:50000"),
				NodeWithAuth("127.0.0.1:50001", "token"),
				WithInsecure(),
				WithRequestTimeout(time.Second),
				WithPingInterval(time.Second),
				WithSuccessCodes(200, 201),
			},
			wantErr: false,
			expectedRes: &writerConfig{
				QueueCap: 10,
				Logger:   log.New(os.Stdout, "", log.Ltime),
				NodeConfigs: []transport.NodeConfig{
					{
						Host:      "127.0.0.1:50000",
						AuthToken: "",
					},
					{
						Host:      "127.0.0.1:50001",
						AuthToken: "token",
					},
				},
				Insecure:       true,
				RequestTimeout: time.Second,
				PingInterval:   time.Second,
				SuccessCodes:   []int{200, 201},
			},
			expectedErr: "",
		},
		{
			name: "Pass#1",
			options: []Option{
				Node("127.0.0.1:50000"),
			},
			wantErr: false,
			expectedRes: &writerConfig{
				QueueCap: DefaultQueueCap,
				NodeConfigs: []transport.NodeConfig{
					{
						Host: "127.0.0.1:50000",
					},
				},
				RequestTimeout: DefaultRequestTimeout,
				PingInterval:   DefaultPingInterval,
				SuccessCodes:   []int{http.StatusOK},
			},
			expectedErr: "",
		},
		{
			name:        "ErrorNoNodesHosts",
			options:     []Option{},
			wantErr:     true,
			expectedErr: ErrNodeHostsIsEmpty.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := buildWriterConfig(tt.options...)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			assert.Equal(t, tt.expectedRes, config)

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}

func TestTransportConfig(t *testing.T) {
	expected := transport.Config{
		NodeConfigs: []transport.NodeConfig{
			{
				Host:      "127.0.0.1:50000",
				AuthToken: "token1",
			},
			{
				Host:      "127.0.0.1:50001",
				AuthToken: "token2",
			},
		},
		RequestTimeout: DefaultRequestTimeout,
		PingInterval:   DefaultPingInterval,
		SuccessCodes:   []int{200, 201, 202},
	}

	config := writerConfig{
		NodeConfigs: []transport.NodeConfig{
			{
				Host:      "127.0.0.1:50000",
				AuthToken: "token1",
			},
			{
				Host:      "127.0.0.1:50001",
				AuthToken: "token2",
			},
		},
		RequestTimeout: DefaultRequestTimeout,
		PingInterval:   DefaultPingInterval,
		SuccessCodes:   []int{200, 201, 202},
	}

	assert.Equal(t, expected, config.transportConfig())
}
