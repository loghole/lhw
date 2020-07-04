package lhw

import (
	"net/http"
	"time"

	"github.com/gadavy/lhw/transport"
)

const (
	// Default writer settings
	DefaultBatchSize    = 1024 * 1024
	DefaultRotatePeriod = time.Second

	// Default transport settings
	DefaultPingInterval   = time.Second
	DefaultRequestTimeout = 2 * time.Second
	DefaultUserAgent      = "go-log-writer"

	// Default storage settings
	DefaultFilepath = "logs/app.log"
)

const (
	MinimalBatchSize = 512
)

type Config struct {
	// Writer settings
	BatchSize    int
	RotatePeriod time.Duration
	Logger       Logger

	// Transport settings
	NodeURIs       []string
	RequestTimeout time.Duration
	PingInterval   time.Duration
	SuccessCodes   []int
	UserAgent      string

	// Storage settings
	Filepath    string
	DropStorage bool
}

func (c *Config) validate() {
	// Check writer settings
	if c.BatchSize <= MinimalBatchSize {
		c.BatchSize = MinimalBatchSize
	}

	// Check transport settings
	if c.RotatePeriod <= 0 {
		c.RotatePeriod = DefaultRotatePeriod
	}

	if c.RequestTimeout <= 0 {
		c.RequestTimeout = DefaultRequestTimeout
	}

	if c.PingInterval <= 0 {
		c.PingInterval = DefaultPingInterval
	}

	if len(c.SuccessCodes) == 0 {
		c.SuccessCodes = []int{
			http.StatusOK,
		}
	}

	if c.UserAgent == "" {
		c.UserAgent = DefaultUserAgent
	}

	// Check storage settings
	if c.Filepath == "" {
		c.Filepath = DefaultFilepath
	}
}

func (c *Config) getTransportConfig() transport.Config {
	return transport.Config{
		NodeURIs:       c.NodeURIs,
		RequestTimeout: c.RequestTimeout,
		PingInterval:   c.PingInterval,
		SuccessCodes:   c.SuccessCodes,
		UserAgent:      c.UserAgent,
	}
}
