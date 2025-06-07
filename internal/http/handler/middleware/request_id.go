package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type RequestID struct {
}

func NewRequestIDMiddleware() *RequestID {
	return &RequestID{}
}

func (r *RequestID) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
