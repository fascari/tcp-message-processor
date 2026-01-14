package integration

import (
	"tcp-message-processor/common/pkg/errlog"
)

func (s *Suite) stopServer() {
	if s.serverCmd == nil || s.serverCmd.Process == nil {
		return
	}

	errlog.Log(s.serverCmd.Process.Kill(), "failed to kill server process")
	errlog.Debug(s.serverCmd.Wait(), "server process wait error")
}

func (s *Suite) stopContainers() {
	if s.postgresContainer != nil {
		errlog.Log(s.postgresContainer.Terminate(s.ctx), "failed to terminate postgres container")
	}

	if s.rabbitmqContainer != nil {
		errlog.Log(s.rabbitmqContainer.Terminate(s.ctx), "failed to terminate rabbitmq container")
	}
}
