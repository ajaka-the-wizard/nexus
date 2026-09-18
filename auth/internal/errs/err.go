package errs

import "errors"

var (
	ERR_DUPLICATE_EMAIL         = errors.New("This email already exists")
	ERR_EMAIL_NO_EXISTS         = errors.New("This email does not exist")
	ERR_INVALID_METHOD          = errors.New("Invalid jwt method")
	ERR_BLACKLISTED_TOKEN       = errors.New("JWT token is blacklisted")
	ERR_BLACKLIST_TOKEN_FAILURE = errors.New("Failed to blacklist JWT token")
	ERR_KEY_DOES_NOT_EXIST      = errors.New("The requested key does not exist")
	ERR_NO_TOKENS_PROVIDED      = errors.New("No tokens were provided in the request")
	ERR_PASSWORD_REUSED         = errors.New("Password was used previously")
	ERR_PASSWORD_RESET_COOLDOWN = errors.New("Password reset is not available yet")
)
