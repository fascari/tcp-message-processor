package errors

import (
	"errors"
)

var (
	ErrTaskNotExist        = errors.New("task does not exist")
	ErrTaskExpired         = errors.New("task expired")
	ErrInvalidResult       = errors.New("invalid result")
	ErrSubmissionFrequent  = errors.New("submission too frequent")
	ErrDuplicateSubmission = errors.New("duplicate submission")
	ErrUnauthorized        = errors.New("unauthorized")
)
