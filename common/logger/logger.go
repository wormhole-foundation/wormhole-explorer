package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Option func(*zap.Config)

func WithLevel(level string) Option {
	return func(c *zap.Config) {
		lvl, err := zap.ParseAtomicLevel(level)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parsing log level '%s'\n", err)
		} else {
			c.Level = lvl
		}
	}
}

func New(system string, opts ...Option) *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	for _, opt := range opts {
		opt(&config)
	}
	l, err := config.Build()
	if err != nil {
		panic(err)
	}
	return l.Named(system)
}
