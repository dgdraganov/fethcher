package repository

import (
	"errors"
	"fethcher/internal/db"
	"fmt"

	"github.com/google/uuid"
)

var ErrUserNotFound error = errors.New("user not found")

type FethRepo struct {
	db Database
}

func NewFethRepo(db Database) *FethRepo {
	return &FethRepo{
		db: db,
	}
}

func (r *FethRepo) MigrateAndSeed(dbName string) error {

	err := r.db.MigrateTable(&Transaction{}, &User{})
	if err != nil {
		return fmt.Errorf("migrate table(s): %w", err)
	}

	users := []User{
		{
			ID:           uuid.NewString(),
			Username:     "alice",
			PasswordHash: "$2a$10$7PrikY/17DYiRAA6JlaGl.yo26gwhTT53ESuovxGWvWJ4HhvGI/GK",
		},
		{
			ID:           uuid.NewString(),
			Username:     "bob",
			PasswordHash: "$2a$10$SHWr22XIYjY3/nLI6QOSJezr5KAB2AUs740F8NahmhBNsPsKacL8u",
		},
		{
			ID:           uuid.NewString(),
			Username:     "carol",
			PasswordHash: "$2a$10$sIVvau/Udc4hgV/xny/IE.LRHVVuTiMF0UTGt.SFfRhCYvunds4h2",
		},
		{
			ID:           uuid.NewString(),
			Username:     "dave",
			PasswordHash: "$2a$10$53qBwnstmYjn4S5HbYoiYe5i.SyQxyZfBiPiCoB1241HRtpVYFMvG",
		},
	}
	err = r.db.SeedDB(&users)
	if err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	return nil
}

func (r *FethRepo) GetUserFromDB(username, password string) (User, error) {
	var user User

	err := r.db.GetBy("username", username, &user)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}
