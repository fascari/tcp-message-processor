//go:build integration

package testsuite

import (
	"net"

	"tcp-message-processor/common/pkg/integration"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type Integration struct {
	suite.Suite
	integration *integration.Suite
}

func (s *Integration) SetupSuite() {
	var err error
	s.integration, err = integration.NewSuite(
		integration.WithDatabase(),
		integration.WithRabbitMQ(),
		integration.WithServer(),
	)
	require.NoError(s.T(), err)
}

func (s *Integration) TearDownSuite() {
	if s.integration != nil {
		s.integration.Teardown()
	}
}

func (s *Integration) Connect() (net.Conn, error) {
	return s.integration.Connect()
}
