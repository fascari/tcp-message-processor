package handler

import (
	"context"
	"net"
	"time"

	"tcp-message-processor/common/pkg/hash"
	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/common/pkg/tcp"
	apperrors "tcp-message-processor/pkg/errors"

	"go.uber.org/zap"
)

type submitParams struct {
	jobID       int64
	clientNonce string
	result      string
}

func (s *Server) parseSubmitParams(conn net.Conn, msg tcp.Message) (submitParams, bool) {
	jobID, ok := msg.Params["job_id"].(float64)
	if !ok {
		s.sendError(conn, *msg.ID, "invalid job_id")
		return submitParams{}, false
	}

	clientNonce, ok := msg.Params["client_nonce"].(string)
	if !ok {
		s.sendError(conn, *msg.ID, "invalid client_nonce")
		return submitParams{}, false
	}

	result, ok := msg.Params["result"].(string)
	if !ok {
		s.sendError(conn, *msg.ID, "invalid result")
		return submitParams{}, false
	}

	return submitParams{
		jobID:       int64(jobID),
		clientNonce: clientNonce,
		result:      result,
	}, true
}

func (s *Server) submit(ctx context.Context, conn net.Conn, msg tcp.Message, username string) {
	params, ok := s.parseSubmitParams(conn, msg)
	if !ok {
		return
	}

	sess, exists := s.sessions.Find(username)
	if !exists {
		s.sendError(conn, *msg.ID, apperrors.ErrUnauthorized.Error())
		return
	}

	if !sess.AllowSubmission() {
		s.sendError(conn, *msg.ID, apperrors.ErrSubmissionFrequent.Error())
		return
	}

	if sess.IsDuplicateNonce(params.clientNonce) {
		s.sendError(conn, *msg.ID, apperrors.ErrDuplicateSubmission.Error())
		return
	}

	if sess.IsJobExpired(params.jobID) {
		s.sendError(conn, *msg.ID, apperrors.ErrTaskExpired.Error())
		return
	}

	if !sess.ValidateJobNonce(params.jobID, sess.CurrentNonce) {
		s.sendError(conn, *msg.ID, apperrors.ErrTaskNotExist.Error())
		return
	}

	expectedResult := hash.SHA256(sess.CurrentNonce + params.clientNonce)
	if params.result != expectedResult {
		logger.Warn("invalid result",
			zap.String("username", username),
			zap.Int64("job_id", params.jobID),
			zap.String("expected", expectedResult),
			zap.String("received", params.result),
		)
		s.sendError(conn, *msg.ID, apperrors.ErrInvalidResult.Error())
		return
	}

	sess.RecordSubmission(params.clientNonce)

	event := Event{
		Username:    username,
		JobID:       params.jobID,
		ClientNonce: params.clientNonce,
		Timestamp:   time.Now(),
	}

	if err := s.publisher.Publish(ctx, event); err != nil {
		logger.Error("failed to publish submission event",
			zap.Error(err),
			zap.String("username", username),
		)
	}

	logger.Info("submission accepted",
		zap.String("username", username),
		zap.Int64("job_id", params.jobID),
		zap.String("client_nonce", params.clientNonce),
	)

	response := tcp.NewSuccessResponse(*msg.ID, true)
	if err := tcp.WriteMessage(conn, response); err != nil {
		logger.Error("failed to send response", zap.Error(err))
	}
}
