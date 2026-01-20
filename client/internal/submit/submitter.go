package submit

import (
	"fmt"
	"time"

	"tcp-message-processor/common/pkg/hash"
	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/common/pkg/noncegen"
	"tcp-message-processor/common/pkg/random"
	"tcp-message-processor/common/pkg/tcp"

	"go.uber.org/zap"
)

type Submitter struct {
	minDelay time.Duration
	maxDelay time.Duration
	idSeq    int
}

func New(minSeconds, maxSeconds int) Submitter {
	return Submitter{
		minDelay: time.Duration(minSeconds) * time.Second,
		maxDelay: time.Duration(maxSeconds) * time.Second,
		idSeq:    2,
	}
}

func (s *Submitter) Submit(conn *tcp.Conn, jobID int64, serverNonce string) error {
	delay := random.Delay(s.minDelay, s.maxDelay)
	time.Sleep(delay)

	clientNonce := noncegen.Generate()
	result := hash.SHA256(serverNonce + clientNonce)

	id := int64(s.idSeq)
	s.idSeq++

	params := tcp.SubmitParams{
		JobID:       jobID,
		ClientNonce: clientNonce,
		Result:      result,
	}
	msg := params.ToMessage(id)

	if err := conn.Write(&msg); err != nil {
		return fmt.Errorf("failed to send submit request: %w", err)
	}

	response, err := conn.Read()
	if err != nil {
		return fmt.Errorf("failed to read submit response: %w", err)
	}

	if response.Error != "" {
		logger.Warn("submission rejected",
			zap.String("error", response.Error),
			zap.Int64("job_id", jobID))
		return nil
	}

	logger.Info("submission accepted",
		zap.Int64("job_id", jobID),
		zap.String("client_nonce", clientNonce))

	return nil
}
