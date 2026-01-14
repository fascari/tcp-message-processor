package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"sync"
	"time"

	"tcp-message-processor/internal/session"
	"tcp-message-processor/pkg/logger"
	"tcp-message-processor/pkg/tcp"

	"go.uber.org/zap"
)

type broadcaster struct {
	mu                sync.RWMutex
	clients           map[string]net.Conn
	jobID             int64
	nonce             string
	ticker            *time.Ticker
	stopChan          chan struct{}
	sessions          *session.Store
	broadcastInterval time.Duration
}

func newBroadcaster(sessions *session.Store, intervalSeconds int) *broadcaster {
	return &broadcaster{
		clients:           make(map[string]net.Conn),
		jobID:             0,
		nonce:             generateNonce(),
		stopChan:          make(chan struct{}),
		sessions:          sessions,
		broadcastInterval: time.Duration(intervalSeconds) * time.Second,
	}
}

func (b *broadcaster) register(username string, conn net.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.clients[username] = conn
	logger.Debug("client registered", zap.String("username", username))
}

func (b *broadcaster) unregister(username string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.clients, username)
	logger.Debug("client unregistered", zap.String("username", username))
}

func (b *broadcaster) start(ctx context.Context) {
	b.ticker = time.NewTicker(b.broadcastInterval)

	go func() {
		for {
			select {
			case <-b.stopChan:
				return
			case <-b.ticker.C:
				b.broadcast(ctx)
			}
		}
	}()

	logger.Info("job broadcaster started", zap.Duration("interval", b.broadcastInterval))
}

func (b *broadcaster) stop() {
	if b.ticker != nil {
		b.ticker.Stop()
	}
	close(b.stopChan)
	logger.Info("job broadcaster stopped")
}

func (b *broadcaster) broadcast(_ context.Context) {
	b.mu.Lock()
	b.jobID++
	b.nonce = generateNonce()
	jobID := b.jobID
	nonce := b.nonce
	clients := make(map[string]net.Conn)
	for k, v := range b.clients {
		clients[k] = v
	}
	b.mu.Unlock()

	params := JobParams{
		JobID:       jobID,
		ServerNonce: nonce,
	}

	msg := tcp.NewNotification(string(MethodJob), params.ToMap())

	logger.Info("broadcasting new job",
		zap.Int64("job_id", jobID),
		zap.String("server_nonce", nonce),
		zap.Int("clients", len(clients)),
	)

	for username, conn := range clients {
		sess, exists := b.sessions.Find(username)
		if !exists {
			continue
		}

		sess.UpdateJob(jobID, nonce)

		if err := tcp.WriteMessage(conn, msg); err != nil {
			logger.Error("failed to send job to client",
				zap.Error(err),
				zap.String("username", username),
			)
		}
	}
}

func generateNonce() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("failed to generate nonce: " + err.Error())
	}
	return hex.EncodeToString(bytes)
}
