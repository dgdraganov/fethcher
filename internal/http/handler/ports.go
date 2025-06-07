package handler

import (
	"context"
	"fethcher/internal/core"
	"net/http"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name RequestValidator . RequestValidator
type RequestValidator interface {
	DecodeJSONPayload(r *http.Request, object any) error
}

//counterfeiter:generate -o fake -fake-name TransactionService . TransactionService
type TransactionService interface {
	Authenticate(ctx context.Context, msg core.AuthMessage) (string, error)
	GetTransactions(ctx context.Context, transactionsHashes []string) ([]core.TransactionRecord, error)
	SaveUserTransactionsHistory(ctx context.Context, token string, transactionsHashes []string) error
	GetUserTransactionsHistory(ctx context.Context, token string) ([]core.TransactionRecord, error)
	GetAllDBTransactions(ctx context.Context) ([]core.TransactionRecord, error)
	ParseRLP(rlphex string) ([]string, error)
}
