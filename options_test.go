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
		name        string
		option      Option
		wantErr     bool
		expectedRes *Options
		expectedErr string
	}{
		{
			name:        "WithQueueCap",
			option:      WithQueueCap(10),
			expectedRes: &Options{QueueCap: 10},
		},
		{
			name:        "WithQueueCapError",
			option:      WithQueueCap(-10),
			wantErr:     true,
			expectedErr: ErrBadQueueCapacity.Error(),
		},
		{
			name:        "WithLogger",
			option:      WithLogger(log.New(os.Stdout, "", log.Ltime)),
			expectedRes: &Options{Logger: log.New(os.Stdout, "", log.Ltime)},
		},
		{
			name:        "WithInsecure",
			option:      WithInsecure(),
			expectedRes: &Options{Insecure: true},
		},
		{
			name:        "WithRequestTimeout",
			option:      WithRequestTimeout(time.Second),
			expectedRes: &Options{RequestTimeout: time.Second},
		},
		{
			name:        "WithRequestTimeoutError",
			option:      WithRequestTimeout(-time.Second),
			wantErr:     true,
			expectedErr: ErrBadRequestTimeout.Error(),
		},
		{
			name:        "WithPingInterval",
			option:      WithPingInterval(time.Second),
			expectedRes: &Options{PingInterval: time.Second},
		},
		{
			name:        "WithPingIntervalError",
			option:      WithPingInterval(-time.Second),
			wantErr:     true,
			expectedErr: ErrBadPingInterval.Error(),
		},
		{
			name:        "WithSuccessCodes",
			option:      WithSuccessCodes(200, 201),
			expectedRes: &Options{SuccessCodes: []int{200, 201}},
		},
		{
			name:        "WithSuccessCodesError",
			option:      WithSuccessCodes(),
			wantErr:     true,
			expectedErr: ErrSuccessCodes.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Options{}

			err := tt.option(config)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			if !tt.wantErr {
				assert.Equal(t, tt.expectedRes, config)
			}

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}

func TestTransportConfig(t *testing.T) {
	expected := transport.Config{
		Servers:        []string{"http://token1@127.0.0.1:50000", "http://token2@127.0.0.1:50001"},
		RequestTimeout: DefaultRequestTimeout,
		PingInterval:   DefaultPingInterval,
		SuccessCodes:   []int{200, 201, 202},
	}

	config := Options{
		Servers:        []string{"http://token1@127.0.0.1:50000", "http://token2@127.0.0.1:50001"},
		RequestTimeout: DefaultRequestTimeout,
		PingInterval:   DefaultPingInterval,
		SuccessCodes:   []int{200, 201, 202},
	}

	assert.Equal(t, expected, config.transportConfig())
}

func TestGetDefaultOptions(t *testing.T) {
	expected := &Options{
		QueueCap:       DefaultQueueCap,
		Insecure:       false,
		RequestTimeout: DefaultRequestTimeout,
		PingInterval:   DefaultPingInterval,
		SuccessCodes:   []int{http.StatusOK, http.StatusCreated},
	}

	assert.Equal(t, expected, GetDefaultOptions())
}