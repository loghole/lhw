package zap

import (
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/loghole/lhw"
)

type Config struct {
	Level         string
	CollectorURL  string
	Hostname      string
	Namespace     string
	Source        string
	BuildCommit   string
	ConfigHash    string
	DisableStdout bool
}

type Option func(options []zap.Option) []zap.Option

func AddCaller() Option {
	return func(options []zap.Option) []zap.Option {
		return append(options, zap.AddCaller())
	}
}

func AddStacktrace(level string) Option {
	return func(options []zap.Option) []zap.Option {
		return append(options, zap.AddStacktrace(zapLevel(level)))
	}
}

func WithField(key string, value interface{}) Option {
	return func(options []zap.Option) []zap.Option {
		return append(options, zap.Fields(zap.Any(key, value)))
	}
}

type Logger struct {
	*zap.SugaredLogger
	closer io.Closer
}

func NewLogger(config *Config, options ...Option) (*Logger, error) {
	logger := &Logger{}

	cores := make([]zapcore.Core, 0)

	if !config.DisableStdout {
		cores = append(cores, consoleCore(config))
	}

	if config.CollectorURL != "" {
		core, closer, err := lhwCore(config)
		if err != nil {
			return nil, err
		}

		logger.closer = closer

		cores = append(cores, core)
	}

	opts := make([]zap.Option, 0, len(options))

	for _, option := range options {
		if option == nil {
			continue
		}

		opts = option(opts)
	}

	logger.SugaredLogger = zap.New(zapcore.NewTee(cores...), opts...).Sugar()

	return logger, nil
}

func (l *Logger) Close() {
	if l.closer != nil {
		_ = l.closer.Close()
	}
}

func zapLevel(lvl string) zapcore.Level {
	switch strings.ToLower(lvl) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "err", "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

func consoleCore(config *Config) zapcore.Core {
	return zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig()), os.Stdout, zapLevel(config.Level))
}

func lhwCore(config *Config) (zapcore.Core, io.Closer, error) {
	writer, err := lhw.NewWriter(config.CollectorURL)
	if err != nil {
		return nil, nil, err
	}

	fields := []zap.Field{zap.String("host", hostname(config.Hostname))}

	if config.Namespace != "" {
		fields = append(fields, zap.String("namespace", config.Namespace))
	}

	if config.Source != "" {
		fields = append(fields, zap.String("source", config.Source))
	}

	if config.BuildCommit != "" {
		fields = append(fields, zap.String("build_commit", config.BuildCommit))
	}

	if config.ConfigHash != "" {
		fields = append(fields, zap.String("config_hash", config.ConfigHash))
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig()), zapcore.AddSync(writer), zapLevel(config.Level))

	return core.With(fields), writer, nil
}

func encoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func hostname(host string) string {
	if host != "" {
		return host
	}

	host, err := os.Hostname()
	if err != nil {
		return "unknown-host"
	}

	return host
}
