package handler

import (
	"context"
	"errors"
	"time"

	"tcp-message-processor/common/pkg/hash"
	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/common/pkg/tcp"
	"tcp-message-processor/internal/events"
	"tcp-message-processor/internal/session"
	apperrors "tcp-message-processor/pkg/errors"

	"go.uber.org/zap"
)

type submitParams struct {
	jobID       int64
	clientNonce string
	result      string
}

func (h *handler) parseSubmitParams(conn *tcp.Conn, msg tcp.Message) (submitParams, bool) {
	jobID, ok := msg.Params["job_id"].(float64)
	if !ok {
		h.sendError(conn, *msg.ID, "invalid job_id")
		return submitParams{}, false
	}

	clientNonce, ok := msg.Params["client_nonce"].(string)
	if !ok {
		h.sendError(conn, *msg.ID, "invalid client_nonce")
		return submitParams{}, false
	}

	result, ok := msg.Params["result"].(string)
	if !ok {
		h.sendError(conn, *msg.ID, "invalid result")
		return submitParams{}, false
	}

	return submitParams{
		jobID:       int64(jobID),
		clientNonce: clientNonce,
		result:      result,
	}, true
}

func (h *handler) submit(ctx context.Context, conn *tcp.Conn, msg tcp.Message, username string) {
	params, ok := h.parseSubmitParams(conn, msg)
	if !ok {
		return
	}

	sess, exists := h.sessions.Find(username)
	if !exists {
		h.sendError(conn, *msg.ID, apperrors.ErrUnauthorized.Error())
		return
	}

	if !sess.AllowSubmission() {
		h.sendError(conn, *msg.ID, apperrors.ErrSubmissionFrequent.Error())
		return
	}

	if sess.IsDuplicateNonce(params.clientNonce) {
		h.sendError(conn, *msg.ID, apperrors.ErrDuplicateSubmission.Error())
		return
	}

	if err := h.validateJobAndResult(conn, msg, sess, params, username); err != nil {
		return
	}

	sess.RecordSubmission(params.clientNonce)

	event := events.Submission{
		Username:    username,
		JobID:       params.jobID,
		ClientNonce: params.clientNonce,
		Timestamp:   time.Now(),
	}

	if err := h.publisher.Publish(ctx, event); err != nil {
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

	response := tcp.SuccessResponse(*msg.ID, true)
	if err := conn.Write(&response); err != nil {
		logger.Error("failed to send response", zap.Error(err))
	}
}

func (h *handler) validateJobAndResult(conn *tcp.Conn, msg tcp.Message, sess *session.Session, params submitParams, username string) error {
	validation := sess.ValidateJob(params.jobID)

	if !validation.Exists {
		h.sendError(conn, *msg.ID, apperrors.ErrTaskNotExist.Error())
		return errors.New("job does not exist")
	}

	if validation.IsExpired {
		h.sendError(conn, *msg.ID, apperrors.ErrTaskExpired.Error())
		return errors.New("job expired")
	}

	if !validation.NonceMatches {
		h.sendError(conn, *msg.ID, apperrors.ErrTaskNotExist.Error())
		return errors.New("nonce mismatch")
	}

	expectedResult := hash.SHA256(validation.CurrentNonce + params.clientNonce)
	if params.result != expectedResult {
		logger.Warn("invalid result",
			zap.String("username", username),
			zap.Int64("job_id", params.jobID),
			zap.String("expected", expectedResult),
			zap.String("received", params.result),
		)
		h.sendError(conn, *msg.ID, apperrors.ErrInvalidResult.Error())
		return errors.New("invalid result")
	}

	return nil
}
