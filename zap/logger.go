package zap

import (
	"github.com/loghole/lhw/zaplog"
)

// Config
// Deprecated: use zaplog.Config.
type Config = zaplog.Config

// Option
// Deprecated: use zaplog.Option.
type Option = zaplog.Option

// AddCaller
// Deprecated: use zaplog.AddCaller.
func AddCaller() Option {
	return zaplog.AddCaller()
}

// AddStacktrace
// Deprecated: use zaplog.AddStacktrace.
func AddStacktrace(level string) Option {
	return zaplog.AddStacktrace(level)
}

// WithField
// Deprecated: use zaplog.WithField.
func WithField(key string, value interface{}) Option {
	return zaplog.WithField(key, value)
}

// Logger
// Deprecated: use zaplog.Logger.
type Logger = zaplog.Logger

// NewLogger
// Deprecated: use zaplog.NewLogger.
func NewLogger(config *Config, options ...Option) (*Logger, error) {
	return zaplog.NewLogger(config, options...)
}
