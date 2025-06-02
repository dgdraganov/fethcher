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
	GetUserFromDB(username, password string) (repository.User, error)
	GetTransactionsByHash(txHashes []string) ([]repository.Transaction, error)
	SaveTransactions(transactions []repository.Transaction) error
	GetUserHistory(userID string) ([]string, error)
	SaveUserHistory(userID string, transactions []string) error
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
