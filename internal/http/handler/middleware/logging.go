package middleware

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Logging struct {
	logs *zap.SugaredLogger
}

func NewLoggingMiddleware(logger *zap.SugaredLogger) *Logging {
	return &Logging{
		logs: logger,
	}
}

func (l *Logging) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestId string = "None"

		req := r.Context().Value(RequestIDKey)
		if req != nil {
			requestId = req.(string)
		}

		start := time.Now()
		msg := fmt.Sprintf("[%s] %s", r.Method, r.RequestURI)
		l.logs.Infow(msg,
			"remote_addr", r.RemoteAddr,
			"request_id", requestId)

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		msg = fmt.Sprintf("[%s] %s", r.Method, r.RequestURI)
		l.logs.Infow(msg,
			"duration", duration,
			"remote_addr", r.RemoteAddr,
			"request_id", requestId)
	})
}
