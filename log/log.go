package log

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates new logger that defined as zap.Logger
func NewLogger(levelStr string) (*zap.Logger, error) {
	level, err := parseLogLevel(levelStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse log level: %s", err.Error())
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		DisableStacktrace: true,
		Development:       false,
		Encoding:          "json",
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	return config.Build()
}

func parseLogLevel(levelStr string) (zapcore.Level, error) {
	var level zapcore.Level

	switch strings.ToUpper(levelStr) {
	case zapcore.DebugLevel.CapitalString():
		level = zapcore.DebugLevel
	case zapcore.InfoLevel.CapitalString():
		level = zapcore.InfoLevel
	case zapcore.WarnLevel.CapitalString():
		level = zapcore.WarnLevel
	case zapcore.ErrorLevel.CapitalString():
		level = zapcore.ErrorLevel
	default:
		return level, fmt.Errorf("Undefined log level: %s", levelStr)
	}

	return level, nil
}
