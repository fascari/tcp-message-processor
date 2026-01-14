package errlog

import (
	"tcp-message-processor/common/pkg/logger"

	"go.uber.org/zap"
)

func Log(err error, msg string) {
	if err != nil {
		logger.Error(msg, zap.Error(err))
	}
}

func Debug(err error, msg string) {
	if err != nil {
		logger.Debug(msg, zap.Error(err))
	}
}
