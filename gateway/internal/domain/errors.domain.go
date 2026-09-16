package domain

import "errors"

var (
	ErrNoSuchService          error = errors.New("The requested service does not exist")
	ErrInvalidMethod          error = errors.New("Invalid Signing method")
	ErrInvalidRedisReturnType error = errors.New("Redis returned an invalid data type")
)
