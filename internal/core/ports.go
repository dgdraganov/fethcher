package core

import (
	"context"
	"fethcher/internal/ethereum"
	"fethcher/internal/repository"
	tokenIssuer "fethcher/pkg/jwt"

	"github.com/golang-jwt/jwt"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name Repository . Repository
type Repository interface {
	GetUserFromDB(ctx context.Context, username, password string) (repository.User, error)
	GetTransactionsByHash(ctx context.Context, txHashes []string) ([]repository.Transaction, error)
	SaveTransactions(ctx context.Context, transactions []repository.Transaction) error
	GetUserHistory(ctx context.Context, userID string) ([]string, error)
	SaveUserHistory(ctx context.Context, userID string, transactions []string) error
}

//counterfeiter:generate -o fake -fake-name JWTIssuer . JWTIssuer
type JWTIssuer interface {
	Generate(data tokenIssuer.TokenInfo) *jwt.Token
	Sign(token *jwt.Token) (string, error)
	Validate(token string) (jwt.MapClaims, error)
}

//counterfeiter:generate -o fake -fake-name EthereumService . EthereumService
type EthereumService interface {
	FetchTransactions(ctx context.Context, hashes []string) ([]*ethereum.Transaction, error)
}
