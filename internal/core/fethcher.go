package core

import (
	"errors"
	"fethcher/internal/storage"
	tokenIssuer "fethcher/pkg/jwt"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var ErrIncorrectPassword error = errors.New("incorrect password")
var ErrUserNotFound error = errors.New("user not found")

type AuthMessage struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Fethcher struct {
	logs      *zap.SugaredLogger
	repo      Repository
	jwtIssuer JWTIssuer
}

func NewFethcher(logger *zap.SugaredLogger, repo Repository, jwt JWTIssuer) *Fethcher {
	return &Fethcher{
		logs:      logger,
		repo:      repo,
		jwtIssuer: jwt,
	}
}

func (f *Fethcher) Authenticate(msg AuthMessage) (string, error) {
	user, err := f.repo.GetUserFromDB(msg.Username, msg.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", ErrUserNotFound
		}

		return "", fmt.Errorf("get user from db: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(msg.Password)); err != nil {
		return "", ErrIncorrectPassword
	}

	tokenInfo := tokenIssuer.TokenInfo{
		UserName:   user.Username,
		Subject:    user.ID,
		Expiration: 24,
	}

	token := f.jwtIssuer.Generate(tokenInfo)

	signed, err := f.jwtIssuer.Sign(token)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signed, nil
}
