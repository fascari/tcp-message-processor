//go:build integration

package integration_test

import (
	"fmt"
	"testing"
	"time"

	"tcp-message-processor-client/internal/transport"
	"tcp-message-processor/common/pkg/closer"
	"tcp-message-processor/common/pkg/hash"
	"tcp-message-processor/common/pkg/integration"
	"tcp-message-processor/common/pkg/tcp"
	apperrors "tcp-message-processor/pkg/errors"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RateLimitSuite struct {
	suite.Suite
	integrationSuite *integration.Suite
}

func TestRateLimitSuite(t *testing.T) {
	suite.Run(t, new(RateLimitSuite))
}

func (s *RateLimitSuite) SetupSuite() {
	s.integrationSuite = integration.NewSuite()
	err := s.integrationSuite.Setup()
	require.NoError(s.T(), err, "failed to setup integration suite")
}

func (s *RateLimitSuite) TearDownSuite() {
	if s.integrationSuite != nil {
		s.integrationSuite.Teardown()
	}
}

func (s *RateLimitSuite) TestRateLimiting() {
	conn, err := transport.Connect(s.integrationSuite.ServerHost, s.integrationSuite.ServerPort)
	s.Require().NoError(err)
	defer closer.Close(conn, "failed to close connection")

	username := fmt.Sprintf("ratelimit-test-%d", time.Now().UnixNano())

	s.authenticate(conn, username)
	jobID, serverNonce := s.waitForJob(conn)

	accepted, rejected := s.attemptSubmissions(conn, jobID, serverNonce)

	s.T().Logf("Results: %d accepted, %d rejected", accepted, rejected)
	s.Greater(rejected, 0)
	s.LessOrEqual(accepted, 2)
}

func (s *RateLimitSuite) authenticate(conn transport.Connection, username string) {
	authMsg := tcp.AuthorizeParams{Username: username}.ToMessage(1)
	s.Require().NoError(conn.Write(authMsg))

	resp, err := conn.Read()
	s.Require().NoError(err)
	s.Require().Empty(resp.Error)
}

func (s *RateLimitSuite) waitForJob(conn transport.Connection) (int64, string) {
	msg, err := conn.Read()
	s.Require().NoError(err)
	s.Require().True(msg.IsJob())

	jobID := int64(msg.Params["job_id"].(float64))
	serverNonce := msg.Params["server_nonce"].(string)

	return jobID, serverNonce
}

func (s *RateLimitSuite) readResponse(conn transport.Connection) tcp.Message {
	for {
		resp, err := conn.Read()
		s.Require().NoError(err)

		if resp.IsJob() {
			continue
		}

		return resp
	}
}

func (s *RateLimitSuite) attemptSubmissions(conn transport.Connection, jobID int64, serverNonce string) (int, int) {
	accepted := 0
	rejected := 0

	s.T().Log("Attempting 5 rapid submissions")

	for i := 0; i < 5; i++ {
		errMsg := s.submitOnce(conn, jobID, serverNonce, i)
		if errMsg == "" {
			accepted++
			s.T().Logf("  Attempt %d: ACCEPTED", i+1)
			continue
		}

		rejected++
		s.T().Logf("  Attempt %d: REJECTED - %s", i+1, errMsg)
		s.Equal(apperrors.ErrSubmissionFrequent.Error(), errMsg)

		if i < 4 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	return accepted, rejected
}

func (s *RateLimitSuite) submitOnce(conn transport.Connection, jobID int64, serverNonce string, attempt int) string {
	nonce := fmt.Sprintf("nonce-%d-%d", time.Now().UnixNano(), attempt)
	result := hash.SHA256(serverNonce + nonce)

	submitMsg := tcp.SubmitParams{
		JobID:       jobID,
		ClientNonce: nonce,
		Result:      result,
	}.ToMessage(int64(attempt + 2))

	s.Require().NoError(conn.Write(submitMsg))

	resp := s.readResponse(conn)
	return resp.Error
}
