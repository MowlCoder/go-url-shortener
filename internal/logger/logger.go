package logger

import (
	"go.uber.org/zap"
)

// Logger for logging app activity.
type Logger struct {
	zapLogger *zap.Logger
}

// Options for setting up logger.
type Options struct {
	Level        string
	IsProduction bool
}

// LogInfo Available log levels
const (
	LogInfo = "INFO"
)

// NewLogger is construction function to create Logger with Options.
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
		zapLogger: zl.WithOptions(zap.AddCallerSkip(1)),
	}, nil
}

// Info put message to standard output.
func (l Logger) Info(msg string) {
	l.zapLogger.Info(msg)
}
