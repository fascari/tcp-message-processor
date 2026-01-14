package jobs

import (
	"sync"

	"tcp-message-processor/common/pkg/logger"

	"go.uber.org/zap"
)

type Manager struct {
	mu          sync.RWMutex
	jobID       int64
	serverNonce string
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Update(jobID int64, serverNonce string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobID = jobID
	m.serverNonce = serverNonce

	logger.Info("job updated",
		zap.Int64("job_id", jobID),
		zap.String("server_nonce", serverNonce))
}

func (m *Manager) Current() (int64, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.jobID, m.serverNonce
}
