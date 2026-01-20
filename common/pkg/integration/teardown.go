package integration

import (
	"context"
	"syscall"
	"time"

	"tcp-message-processor/common/pkg/errlog"
)

func (s *Suite) stopServer() {
	if s.serverCmd == nil || s.serverCmd.Process == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.signalProcessGroup(syscall.SIGTERM); err != nil {
		errlog.Log(err, "failed to send SIGTERM")
	}

	done := make(chan error, 1)
	go func() {
		done <- s.serverCmd.Wait()
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		if err := s.signalProcessGroup(syscall.SIGKILL); err != nil {
			errlog.Log(err, "failed to send SIGKILL")
		}
		<-done
	}
}

func (s *Suite) signalProcessGroup(sig syscall.Signal) error {
	pgid := s.serverCmd.Process.Pid
	return syscall.Kill(-pgid, sig)
}

func (s *Suite) stopContainers() {
	if s.postgresContainer != nil {
		errlog.Log(s.postgresContainer.Terminate(s.ctx), "failed to terminate postgres container")
	}

	if s.rabbitmqContainer != nil {
		errlog.Log(s.rabbitmqContainer.Terminate(s.ctx), "failed to terminate rabbitmq container")
	}
}
