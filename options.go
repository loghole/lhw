package lhw

import (
	"errors"
	"net/http"
	"time"

	"github.com/loghole/lhw/transport"
)

const (
	DefaultQueueCap       = 1000
	DefaultPingInterval   = time.Second
	DefaultRequestTimeout = 2 * time.Second
)

var (
	ErrBadQueueCapacity  = errors.New("queue capacity invalid")
	ErrBadRequestTimeout = errors.New("request timeout invalid")
	ErrBadPingInterval   = errors.New("ping interval invalid")
	ErrSuccessCodes      = errors.New("success codes empty")
)

type Option func(option *Options) error

func WithQueueCap(capacity int) Option {
	return func(options *Options) error {
		if capacity <= 0 {
			return ErrBadQueueCapacity
		}

		options.QueueCap = capacity

		return nil
	}
}

func WithLogger(logger Logger) Option {
	return func(options *Options) error {
		options.Logger = logger

		return nil
	}
}

func WithInsecure() Option {
	return func(options *Options) error {
		options.Insecure = true

		return nil
	}
}

func WithRequestTimeout(timeout time.Duration) Option {
	return func(options *Options) error {
		if timeout <= 0 {
			return ErrBadRequestTimeout
		}

		options.RequestTimeout = timeout

		return nil
	}
}

func WithPingInterval(interval time.Duration) Option {
	return func(options *Options) error {
		if interval <= 0 {
			return ErrBadPingInterval
		}

		options.PingInterval = interval

		return nil
	}
}

func WithSuccessCodes(codes ...int) Option {
	return func(options *Options) error {
		if len(codes) == 0 {
			return ErrSuccessCodes
		}

		options.SuccessCodes = codes

		return nil
	}
}

type Options struct {
	// Writer settings
	QueueCap int
	Logger   Logger

	Servers        []string
	Insecure       bool
	RequestTimeout time.Duration
	PingInterval   time.Duration
	SuccessCodes   []int
}

// GetDefaultOptions returns default configuration options for the client.
func GetDefaultOptions() *Options {
	return &Options{
		QueueCap:       DefaultQueueCap,
		Insecure:       false,
		RequestTimeout: DefaultRequestTimeout,
		PingInterval:   DefaultPingInterval,
		SuccessCodes:   []int{http.StatusOK, http.StatusCreated},
	}
}

func (o *Options) transportConfig() transport.Config {
	return transport.Config{
		Servers:        o.Servers,
		Insecure:       o.Insecure,
		RequestTimeout: o.RequestTimeout,
		PingInterval:   o.PingInterval,
		SuccessCodes:   o.SuccessCodes,
	}
}
