package stats

import (
	"context"
	"database/sql"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Increment(ctx context.Context, username string, timestamp time.Time) error {
	minuteTimestamp := timestamp.Truncate(time.Minute)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO submissions (username, timestamp, submission_count) 
		VALUES ($1, $2, 1)
		ON CONFLICT (username, timestamp) 
		DO UPDATE SET submission_count = submissions.submission_count + 1
	`, username, minuteTimestamp)

	return err
}
