//go:build integration

package integration_test

import (
	"fmt"
	"testing"
	"time"

	"tcp-message-processor-client/testsuite"
	"tcp-message-processor/common/pkg/closer"
	"tcp-message-processor/common/pkg/hash"
	"tcp-message-processor/common/pkg/tcp"
	apperrors "tcp-message-processor/pkg/errors"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type RateLimitSuite struct {
	testsuite.Integration
}

func TestRateLimitSuite(t *testing.T) {
	suite.Run(t, new(RateLimitSuite))
}

func (s *RateLimitSuite) TestRateLimiting() {
	netConn, err := s.Connect()
	s.Require().NoError(err)
	defer closer.Close(netConn, "failed to close connection")

	conn := tcp.NewConn(netConn)
	s.authenticate(conn, fmt.Sprintf("ratelimit-test-%d", time.Now().UnixNano()))

	jobID, serverNonce := s.waitForJob(conn)
	accepted, rejected := s.attemptSubmissions(conn, jobID, serverNonce)

	s.T().Logf("Rate limit test: %d accepted, %d rejected", accepted, rejected)
	s.Greater(rejected, 0)
	s.LessOrEqual(accepted, 2)
}

func (s *RateLimitSuite) authenticate(conn *tcp.Conn, username string) {
	msg := tcp.AuthorizeParams{Username: username}.ToMessage(1)
	s.Require().NoError(conn.Write(&msg))

	resp, err := conn.Read()
	s.Require().NoError(err)
	s.Require().Empty(resp.Error)
}

func (s *RateLimitSuite) waitForJob(conn *tcp.Conn) (int64, string) {
	msg, err := conn.Read()
	s.Require().NoError(err)
	s.Require().True(msg.IsJob())

	jobID := int64(msg.Params["job_id"].(float64))
	serverNonce := msg.Params["server_nonce"].(string)

	return jobID, serverNonce
}

func (s *RateLimitSuite) readResponse(conn *tcp.Conn) tcp.Message {
	for {
		resp, err := conn.Read()
		s.Require().NoError(err)

		if resp.IsJob() {
			continue
		}

		return resp
	}
}

func (s *RateLimitSuite) attemptSubmissions(conn *tcp.Conn, jobID int64, serverNonce string) (int, int) {
	accepted, rejected := 0, 0

	for i := 0; i < 5; i++ {
		if s.submitOnce(conn, jobID, serverNonce, i) {
			accepted++
		} else {
			rejected++
			if i < 4 {
				time.Sleep(200 * time.Millisecond)
			}
		}
	}

	return accepted, rejected
}

func (s *RateLimitSuite) submitOnce(conn *tcp.Conn, jobID int64, serverNonce string, attempt int) bool {
	nonce := fmt.Sprintf("nonce-%d-%d", time.Now().UnixNano(), attempt)
	result := hash.SHA256(serverNonce + nonce)

	submitMsg := tcp.SubmitParams{
		JobID:       jobID,
		ClientNonce: nonce,
		Result:      result,
	}.ToMessage(int64(attempt + 2))

	s.Require().NoError(conn.Write(&submitMsg))

	resp := s.readResponse(conn)
	if resp.Error != "" {
		s.Equal(apperrors.ErrSubmissionFrequent.Error(), resp.Error)
		return false
	}
	return true
}
