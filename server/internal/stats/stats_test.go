//go:build integration

package stats

import (
	"context"
	"sync"
	"testing"
	"time"

	"tcp-message-processor/testsuite"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StatsTestSuite struct {
	testsuite.Integration
	store *Store
}

func TestStatsTestSuite(t *testing.T) {
	suite.Run(t, new(StatsTestSuite))
}

func (s *StatsTestSuite) SetupSuite() {
	s.Integration.SetupSuite()
	s.store = NewStore(s.DB())
}

func (s *StatsTestSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.DB().ExecContext(ctx, "DELETE FROM submissions")
	s.Require().NoError(err)
}

func (s *StatsTestSuite) TestIncrement_should_create_new_record_when_not_exists() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username := "testuser"
	timestamp := time.Now()

	err := s.store.Increment(ctx, username, timestamp)
	s.Require().NoError(err)

	count := s.submissionCount(username, timestamp)
	s.Require().Equal(1, count)
}

func (s *StatsTestSuite) TestIncrement_should_increment_existing_record() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username := "testuser"
	timestamp := time.Now()

	err := s.store.Increment(ctx, username, timestamp)
	s.Require().NoError(err)

	err = s.store.Increment(ctx, username, timestamp)
	s.Require().NoError(err)

	count := s.submissionCount(username, timestamp)
	s.Require().Equal(2, count)
}

func (s *StatsTestSuite) TestIncrement_should_truncate_to_minute() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username := "testuser"

	t1 := time.Date(2026, 1, 15, 10, 30, 15, 0, time.UTC)
	t2 := time.Date(2026, 1, 15, 10, 30, 45, 0, time.UTC)

	err := s.store.Increment(ctx, username, t1)
	s.Require().NoError(err)

	err = s.store.Increment(ctx, username, t2)
	s.Require().NoError(err)

	count := s.submissionCount(username, t1)
	s.Require().Equal(2, count)
}

func (s *StatsTestSuite) TestIncrement_should_handle_concurrent_writes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	username := "concurrent_user"
	timestamp := time.Now()
	concurrency := 10

	var wg sync.WaitGroup
	errChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.store.Increment(ctx, username, timestamp); err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		s.Require().NoError(err)
	}

	count := s.submissionCount(username, timestamp)
	s.Require().Equal(concurrency, count)
}

func (s *StatsTestSuite) TestIncrement_should_separate_different_users() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	timestamp := time.Now()

	err := s.store.Increment(ctx, "user1", timestamp)
	s.Require().NoError(err)

	err = s.store.Increment(ctx, "user2", timestamp)
	s.Require().NoError(err)

	count1 := s.submissionCount("user1", timestamp)
	count2 := s.submissionCount("user2", timestamp)

	s.Require().Equal(1, count1)
	s.Require().Equal(1, count2)
}

func (s *StatsTestSuite) TestIncrement_should_separate_different_minutes() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username := "testuser"

	t1 := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 15, 10, 31, 0, 0, time.UTC)

	err := s.store.Increment(ctx, username, t1)
	s.Require().NoError(err)

	err = s.store.Increment(ctx, username, t2)
	s.Require().NoError(err)

	count1 := s.submissionCount(username, t1)
	count2 := s.submissionCount(username, t2)

	s.Require().Equal(1, count1)
	s.Require().Equal(1, count2)
}

func (s *StatsTestSuite) submissionCount(username string, timestamp time.Time) int {
	minuteTimestamp := timestamp.Truncate(time.Minute)

	var count int
	err := s.DB().QueryRow(`
		SELECT submission_count 
		FROM submissions 
		WHERE username = $1 AND timestamp = $2
	`, username, minuteTimestamp).Scan(&count)

	require.NoError(s.T(), err)
	return count
}
