package domain

// Service constants
const (
	AUTH_ROUTES string = "auth_service"
	USER_ROUTES string = "user_service"
)

// Application constants

const (
	USER_JWT_COOKIE_KEY string = "JWT_SECRET"
)

// Rate limiting constants

const (
	REFILL_RATE_PER_SECOND uint = 1
	BUCKET_SIZE            uint = 60
)
