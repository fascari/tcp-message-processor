package broadcaster

import (
	"net"
	"sync"
	"time"

	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/common/pkg/noncegen"
	"tcp-message-processor/common/pkg/tcp"
	"tcp-message-processor/internal/session"

	"go.uber.org/zap"
)

type Broadcaster struct {
	mu                sync.RWMutex
	clients           map[string]net.Conn
	jobID             int64
	srvNonce          string
	ticker            *time.Ticker
	stopChan          chan struct{}
	sessions          *session.Store
	broadcastInterval time.Duration
}

func New(sessions *session.Store, intervalSeconds int) Broadcaster {
	return Broadcaster{
		clients:           make(map[string]net.Conn),
		jobID:             0,
		srvNonce:          noncegen.Generate(),
		stopChan:          make(chan struct{}),
		sessions:          sessions,
		broadcastInterval: time.Duration(intervalSeconds) * time.Second,
	}
}

func (b *Broadcaster) Register(username string, conn net.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.clients[username] = conn
	logger.Debug("client registered", zap.String("username", username))
}

func (b *Broadcaster) Unregister(username string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.clients, username)
	logger.Debug("client unregistered", zap.String("username", username))
}

func (b *Broadcaster) Start() {
	b.ticker = time.NewTicker(b.broadcastInterval)

	go func() {
		for {
			select {
			case <-b.stopChan:
				return
			case <-b.ticker.C:
				b.broadcast()
			}
		}
	}()

	logger.Info("job broadcaster started", zap.Duration("interval", b.broadcastInterval))
}

func (b *Broadcaster) Stop() {
	if b.ticker != nil {
		b.ticker.Stop()
	}
	close(b.stopChan)
	logger.Info("job broadcaster stopped")
}

func (b *Broadcaster) broadcast() {
	jobID, serverNonce, clients := b.generateJob()
	msg := createJobMessage(jobID, serverNonce)

	logger.Info("broadcasting new job",
		zap.Int64("job_id", jobID),
		zap.String("server_nonce", serverNonce),
		zap.Int("clients", len(clients)),
	)

	b.sendToClients(clients, msg, jobID, serverNonce)
}

func (b *Broadcaster) generateJob() (int64, string, map[string]net.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.jobID++
	b.srvNonce = noncegen.Generate()

	clients := make(map[string]net.Conn, len(b.clients))
	for k, v := range b.clients {
		clients[k] = v
	}

	return b.jobID, b.srvNonce, clients
}

func createJobMessage(jobID int64, nonce string) tcp.Message {
	params := tcp.JobParams{
		JobID:       jobID,
		ServerNonce: nonce,
	}
	return params.ToMessage()
}

func (b *Broadcaster) sendToClients(clients map[string]net.Conn, msg tcp.Message, jobID int64, srvNonce string) {
	for username, conn := range clients {
		b.sendToClient(username, conn, msg, jobID, srvNonce)
	}
}

func (b *Broadcaster) sendToClient(username string, conn net.Conn, msg tcp.Message, jobID int64, srvNonce string) {
	sess, exists := b.sessions.Find(username)
	if !exists {
		return
	}

	sess.UpdateJob(jobID, srvNonce)

	if err := tcp.WriteMessage(conn, &msg); err != nil {
		logger.Error("failed to send job to client",
			zap.Error(err),
			zap.String("username", username),
		)
	}
}
