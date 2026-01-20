package events

import "time"

type Submission struct {
	Username    string
	JobID       int64
	ClientNonce string
	Timestamp   time.Time
}
