package handler

import (
	"fethcher/internal/core"
	"net/http"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name RequestValidator . RequestValidator
type RequestValidator interface {
	DecodeAndValidateJSONPayload(r *http.Request, object any) error
}

//counterfeiter:generate -o fake -fake-name TransactionService . TransactionService
type TransactionService interface {
	Authenticate(msg core.AuthMessage) (string, error)
}
