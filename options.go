package lhw

import (
	"errors"
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

var (
	ErrNodeHostsIsEmpty = errors.New("no node hosts to connect")
)

type Option func(cfg *writerConfig)

func WithQueueCap(capacity int) Option {
	return func(cfg *writerConfig) {
		cfg.QueueCap = capacity
	}
}

func WithLogger(logger Logger) Option {
	return func(cfg *writerConfig) {
		cfg.Logger = logger
	}
}

func Node(host string) Option {
	return func(cfg *writerConfig) {
		if host != "" {
			cfg.NodeConfigs = append(cfg.NodeConfigs, transport.NodeConfig{Host: host})
		}
	}
}

func NodeWithAuth(host, token string) Option {
	return func(cfg *writerConfig) {
		if host != "" {
			cfg.NodeConfigs = append(cfg.NodeConfigs, transport.NodeConfig{Host: host, AuthToken: token})
		}
	}
}

func WithInsecure() Option {
	return func(cfg *writerConfig) {
		cfg.Insecure = true
	}
}

func WithRequestTimeout(timeout time.Duration) Option {
	return func(cfg *writerConfig) {
		cfg.RequestTimeout = timeout
	}
}

func WithPingInterval(interval time.Duration) Option {
	return func(cfg *writerConfig) {
		cfg.PingInterval = interval
	}
}

func WithSuccessCodes(codes []int) Option {
	return func(cfg *writerConfig) {
		cfg.SuccessCodes = codes
	}
}

type writerConfig struct {
	// Writer settings
	QueueCap int
	Logger   Logger

	// Transport settings
	NodeConfigs    []transport.NodeConfig
	Insecure       bool
	RequestTimeout time.Duration
	PingInterval   time.Duration
	SuccessCodes   []int
}

func buildWriterConfig(options ...Option) (*writerConfig, error) {
	config := &writerConfig{}

	for _, option := range options {
		option(config)
	}

	if config.QueueCap <= 0 {
		config.QueueCap = DefaultQueueCap
	}

	if config.RequestTimeout == 0 {
		config.RequestTimeout = DefaultRequestTimeout
	}

	if config.PingInterval == 0 {
		config.PingInterval = DefaultPingInterval
	}

	if len(config.SuccessCodes) == 0 {
		config.SuccessCodes = []int{
			http.StatusOK,
		}
	}

	if len(config.NodeConfigs) == 0 {
		return nil, ErrNodeHostsIsEmpty
	}

	return config, nil
}

func (c *writerConfig) transportConfig() transport.Config {
	return transport.Config{
		Nodes:          c.NodeConfigs,
		Insecure:       c.Insecure,
		RequestTimeout: c.RequestTimeout,
		PingInterval:   c.PingInterval,
		SuccessCodes:   c.SuccessCodes,
	}
}
