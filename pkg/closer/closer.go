package closer

import (
	"io"

	"tcp-message-processor/pkg/logger"

	"go.uber.org/zap"
)

func Close(c io.Closer, msg string) {
	if c == nil {
		return
	}
	if err := c.Close(); err != nil {
		logger.Error(msg, zap.Error(err))
	}
}
