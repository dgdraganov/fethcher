package core

import (
	"fethcher/internal/repository"
	tokenIssuer "fethcher/pkg/jwt"

	"github.com/golang-jwt/jwt"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fake -fake-name Repository . Repository
type Repository interface {
	GetUserFromDB(username, password string) (repository.User, error)
}

//counterfeiter:generate -o fake -fake-name JWTIssuer . JWTIssuer
type JWTIssuer interface {
	Generate(data tokenIssuer.TokenInfo) *jwt.Token
	Sign(token *jwt.Token) (string, error)
}
