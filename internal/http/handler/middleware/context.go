package middleware

type ctxKey string

const (
	RequestIDKey ctxKey = "request_id"
	AuthTokenKey ctxKey = "auth_token"
)
