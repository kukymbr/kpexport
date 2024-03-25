package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewDefaultConsoleLogger(isDebug bool) *zap.Logger {
	minLevel := zap.InfoLevel
	if isDebug {
		minLevel = zap.DebugLevel
	}

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lvl >= minLevel
	})

	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	encoderConf := zap.NewProductionEncoderConfig()
	encoderConf.EncodeTime = zapcore.TimeEncoderOfLayout("20060102 15:04:05")
	encoderConf.EncodeLevel = zapcore.CapitalColorLevelEncoder

	jsonEncoder := zapcore.NewConsoleEncoder(encoderConf)

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, consoleErrors, highPriority),
		zapcore.NewCore(jsonEncoder, consoleDebugging, lowPriority),
	)

	return zap.New(core)
}
