package domain

import "errors"

var (
	ErrNoSuchService error = errors.New("The requested service does not exist")
)
