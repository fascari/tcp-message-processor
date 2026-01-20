package app

import (
	"context"
	"fmt"

	"tcp-message-processor-client/internal/auth"
	"tcp-message-processor-client/internal/config"
	"tcp-message-processor-client/internal/jobs"
	"tcp-message-processor-client/internal/submit"
	"tcp-message-processor-client/internal/transport"
	"tcp-message-processor/common/pkg/closer"
	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/common/pkg/tcp"

	"go.uber.org/zap"
)

type Client struct {
	conn      *tcp.Conn
	auth      auth.Authenticator
	jobs      *jobs.Manager
	submitter submit.Submitter
}

func New(cfg config.Config) (Client, error) {
	conn, err := transport.Connect(cfg.Server.Host, cfg.Server.Port)
	if err != nil {
		return Client{}, err
	}

	logger.Info("connected to server",
		zap.String("server", fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)))

	authenticator := auth.New(cfg.Client.Username)
	if err := authenticator.Authorize(conn); err != nil {
		closer.Close(conn, "failed to close connection during auth error")
		return Client{}, err
	}

	return Client{
		conn:      conn,
		auth:      authenticator,
		jobs:      jobs.NewManager(),
		submitter: submit.New(cfg.Client.SubmissionMinSeconds, cfg.Client.SubmissionMaxSeconds),
	}, nil
}

func (c *Client) Run(ctx context.Context) error {
	defer closer.Close(c.conn, "failed to close connection")

	return c.handleMessages(ctx)
}

func (c *Client) handleMessages(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg, err := c.conn.Read()
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		if msg.IsJob() {
			c.handleJob(msg)
		}
	}
}

func (c *Client) handleJob(msg tcp.Message) {
	jobID, ok := msg.Params["job_id"].(float64)
	if !ok {
		logger.Error("invalid job_id type")
		return
	}

	serverNonce, ok := msg.Params["server_nonce"].(string)
	if !ok {
		logger.Error("invalid server_nonce type")
		return
	}

	c.jobs.Update(int64(jobID), serverNonce)

	capturedJobID := int64(jobID)
	capturedNonce := serverNonce

	go func() {
		if err := c.submitter.Submit(c.conn, capturedJobID, capturedNonce); err != nil {
			logger.Error("submission failed", zap.Error(err))
		}
	}()
}
