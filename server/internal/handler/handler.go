package handler

import (
	"context"
	"net"

	"tcp-message-processor/common/pkg/closer"
	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/common/pkg/method"
	"tcp-message-processor/common/pkg/tcp"
	"tcp-message-processor/internal/broadcaster"
	"tcp-message-processor/internal/events"
	"tcp-message-processor/internal/session"

	"go.uber.org/zap"
)

//go:generate mockery --all

type (
	Handler interface {
		Handle(ctx context.Context, conn net.Conn)
		Start()
		Stop()
	}

	SessionRepository interface {
		Create(username string) *session.Session
		Find(username string) (*session.Session, bool)
		All() []*session.Session
	}

	Publisher interface {
		Publish(ctx context.Context, event events.Submission) error
	}

	handler struct {
		sessions    SessionRepository
		publisher   Publisher
		broadcaster broadcaster.Broadcaster
	}
)

func New(sessions SessionRepository, publisher Publisher, broadcastInterval int) Handler {
	return &handler{
		sessions:    sessions,
		publisher:   publisher,
		broadcaster: broadcaster.New(sessions, broadcastInterval),
	}
}

func (h *handler) Start() {
	h.broadcaster.Start()
}

func (h *handler) Stop() {
	h.broadcaster.Stop()
}

func (h *handler) Handle(ctx context.Context, conn net.Conn) {
	defer closer.Close(conn, "failed to close connection")

	logger.Info("new connection accepted", zap.String("remote_addr", conn.RemoteAddr().String()))

	tcpConn := tcp.NewConn(conn)
	var username string
	defer func() {
		if username != "" {
			h.broadcaster.Unregister(username)
			logger.Info("client disconnected", zap.String("username", username))
		}
	}()

	for {
		msg, err := tcpConn.Read()
		if err != nil {
			if username != "" {
				logger.Warn("failed to read message", zap.Error(err), zap.String("username", username))
			}
			return
		}

		if msg.ID == nil {
			continue
		}

		if h.processMessage(ctx, tcpConn, msg, &username) {
			return
		}
	}
}

func (h *handler) processMessage(ctx context.Context, conn *tcp.Conn, msg tcp.Message, username *string) bool {
	switch method.Method(msg.Method) {
	case method.Authorize:
		return h.handleAuthorizeMessage(conn, msg, username)
	case method.Submit:
		return h.handleSubmitMessage(ctx, conn, msg, *username)
	default:
		h.sendError(conn, *msg.ID, "unknown method")
		return false
	}
}

func (h *handler) handleAuthorizeMessage(conn *tcp.Conn, msg tcp.Message, username *string) bool {
	user, err := h.authorize(conn, msg)
	if err != nil {
		logger.Error("authorization failed", zap.Error(err))
		return true
	}
	*username = user
	return false
}

func (h *handler) handleSubmitMessage(ctx context.Context, conn *tcp.Conn, msg tcp.Message, username string) bool {
	if username == "" {
		h.sendError(conn, *msg.ID, "unauthorized")
		return true
	}
	h.submit(ctx, conn, msg, username)
	return false
}

func (*handler) sendError(conn *tcp.Conn, id int64, errMsg string) {
	response := tcp.ErrorResponse(id, errMsg)
	if err := conn.Write(&response); err != nil {
		logger.Error("failed to send error response", zap.Error(err))
	}
}
