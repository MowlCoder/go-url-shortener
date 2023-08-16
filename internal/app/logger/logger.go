package logger

import (
	"go.uber.org/zap"
)

type Logger struct {
	zapLogger *zap.Logger
}

type Options struct {
	Level        string
	IsProduction bool
}

const (
	LogInfo = "INFO"
)

func NewLogger(options Options) (*Logger, error) {
	atomicLevel, err := zap.ParseAtomicLevel(options.Level)

	if err != nil {
		return nil, err
	}

	cfg := zap.NewDevelopmentConfig()

	if options.IsProduction {
		cfg = zap.NewProductionConfig()
	}

	cfg.Level = atomicLevel

	zl, err := cfg.Build()

	if err != nil {
		return nil, err
	}

	return &Logger{
		zapLogger: zl,
	}, nil
}

func (l Logger) Info(msg string) {
	l.zapLogger.Info(msg)
}
