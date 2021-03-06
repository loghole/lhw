package main

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gadavy/lhw"
)

func main() {
	// Init writer.
	writer, err := lhw.NewWriter("https://secret_token_1@127.0.0.1:50000",
		lhw.WithLogger(log.New(os.Stdout, "", log.Ldate)),
	)
	if err != nil {
		panic(err)
	}

	// Close flushes any buffered log entries.
	defer writer.Close()

	// init logger
	logger := zap.New(zapcore.NewCore(
		getEncoder(),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(writer), os.Stdout),
		zapcore.DebugLevel)).Sugar()

	// init logger fields
	logger = logger.With("namespace", "example")
	logger = logger.With("source", "example_app")
	logger = logger.With("host", "http://127.0.0.1:8080")
	logger = logger.With("trace_id", "example trace_id")
	logger = logger.With("build_commit", "example build_commit")
	logger = logger.With("config_hash", "example config_hash")

	// example messages
	logger.Debug("debug message")
	logger.Debugf("debug message with %s", "arg")

	logger.Info("info message")
	logger.Infof("info message with %s", "arg")

	logger.Warn("warn message")
	logger.Warnf("warn message with %s", "arg")

	logger.Error("error message")
	logger.Errorf("error message with %s", "arg")
}

// getEncoder return log hole json encoding
func getEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
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
	})
}
