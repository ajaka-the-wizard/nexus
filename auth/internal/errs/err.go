package errs

import "errors"

var (
	ERR_DUPLICATE_EMAIL    = errors.New("This email already exists")
	ERR_EMAIL_NO_EXISTS    = errors.New("This email does not exist")
	ERR_INVALID_METHOD     = errors.New("Invalid jwt method")
	ERR_BLACKLISTED_TOKEN  = errors.New("JWT token is blacklisted")
	ERR_KEY_DOES_NOT_EXIST = errors.New("The requested key does not exist")
)
