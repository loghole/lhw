package lhw

import (
	"net/http"
	"time"

	"github.com/gadavy/lhw/transport"
)

const (
	// Default writer config
	DefaultQueueCap = 1000

	// Default transport settings
	DefaultPingInterval   = time.Second
	DefaultRequestTimeout = 2 * time.Second
)

type Config struct {
	// Writer settings
	QueueCap int
	Logger   Logger

	// Transport settings
	NodeURIs       []string
	Insecure       bool
	RequestTimeout time.Duration
	PingInterval   time.Duration
	SuccessCodes   []int
}

func (c *Config) validate() {
	if c.QueueCap <= 0 {
		c.QueueCap = DefaultQueueCap
	}

	if c.RequestTimeout == 0 {
		c.RequestTimeout = DefaultRequestTimeout
	}

	if c.PingInterval == 0 {
		c.PingInterval = DefaultPingInterval
	}

	if len(c.SuccessCodes) == 0 {
		c.SuccessCodes = []int{
			http.StatusOK,
		}
	}
}

func (c *Config) transportConfig() transport.Config {
	return transport.Config{
		NodeURIs:       c.NodeURIs,
		Insecure:       c.Insecure,
		RequestTimeout: c.RequestTimeout,
		PingInterval:   c.PingInterval,
		SuccessCodes:   c.SuccessCodes,
	}
}
