package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

func Init() error {
	var config zap.Config

	config = zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.DisableCaller = true
	config.DisableStacktrace = true

	logger, err := config.Build()
	if err != nil {
		return err
	}

	globalLogger = logger
	return nil
}

func Log() *zap.Logger {
	if globalLogger == nil {
		return zap.NewNop()
	}
	return globalLogger
}

func Debug(msg string, fields ...zap.Field) {
	Log().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Log().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log().Error(msg, fields...)
}

func Sync() {
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
}
