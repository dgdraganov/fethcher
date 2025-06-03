package storage

import (
	"errors"
	"fethcher/internal/db"
	"fethcher/internal/storage/models"
	"fmt"

	"github.com/google/uuid"
)

var ErrUserNotFound error = errors.New("user not found")

type UserRepository struct {
	db Database
}

func NewUserRepository(db Database) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) MigrateAndSeed(dbName string) error {
	err := r.db.MigrateModels(&models.Transaction{}, &models.User{})
	if err != nil {
		return fmt.Errorf("migrate table(s): %w", err)
	}

	users := []models.User{
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
	err = r.db.Seed(&users)
	if err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUserFromDB(username, password string) (*models.User, error) {
	var user models.User

	err := r.db.GetBy("username", username, &user)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return &user, nil
}
