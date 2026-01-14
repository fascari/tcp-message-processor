package stats

import (
	"context"
	"database/sql"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return Store{db: db}
}

func (s Store) Increment(ctx context.Context, username string, timestamp time.Time) error {
	minuteTimestamp := timestamp.Truncate(time.Minute)

	result, err := s.db.ExecContext(ctx, `
		UPDATE submissions 
		SET submission_count = submission_count + 1 
		WHERE username = $1 AND timestamp = $2
	`, username, minuteTimestamp)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO submissions (username, timestamp, submission_count) 
			VALUES ($1, $2, 1)
		`, username, minuteTimestamp)
		return err
	}

	return nil
}
