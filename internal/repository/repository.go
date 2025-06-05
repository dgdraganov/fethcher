package repository

import (
	"context"
	"errors"
	"fethcher/internal/db"
	"fmt"

	"github.com/google/uuid"
)

var ErrUserNotFound error = errors.New("user not found")

type TransactionRepository struct {
	db Storage
}

func NewTransactionRepository(db Storage) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

func (r *TransactionRepository) MigrateAndSeed() error {

	err := r.db.MigrateTable(&Transaction{}, &User{}, &UserTransaction{})
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
	err = r.db.SaveToTable(context.Background(), &users)
	if err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	return nil
}

func (r *TransactionRepository) SaveTransactions(ctx context.Context, transactions []Transaction) error {
	err := r.db.SaveToTable(ctx, &transactions)
	if err != nil {
		return fmt.Errorf("save to table: %w", err)
	}

	return nil
}
func (r *TransactionRepository) GetUserHistory(ctx context.Context, userID string) ([]string, error) {
	var userTransactions []UserTransaction

	err := r.db.GetAllBy(ctx, "user_id", userID, &userTransactions)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, fmt.Errorf("get user history: %w", ErrUserNotFound)
		}
		return nil, fmt.Errorf("get user history: %w", err)
	}

	txHashes := make([]string, 0, len(userTransactions))
	for _, tx := range userTransactions {
		txHashes = append(txHashes, tx.TransactionHash)
	}

	return txHashes, nil
}

func (r *TransactionRepository) SaveUserHistory(ctx context.Context, userID string, transactions []string) error {
	if len(transactions) == 0 {
		return nil
	}

	userTransactions := make([]UserTransaction, 0, len(transactions))
	for _, tx := range transactions {
		userTransactions = append(userTransactions, UserTransaction{
			UserID:          userID,
			TransactionHash: tx,
		})
	}

	err := r.db.SaveToTable(ctx, &userTransactions)
	if err != nil {
		return fmt.Errorf("save user history: %w", err)
	}

	return nil
}

func (r *TransactionRepository) GetUserFromDB(ctx context.Context, username, password string) (User, error) {
	var user User

	err := r.db.GetOneBy(ctx, "username", username, &user)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}

func (r *TransactionRepository) GetTransactionsByHash(ctx context.Context, txHashes []string) ([]Transaction, error) {
	transactions := []Transaction{}
	err := r.db.GetAllBy(ctx, "transaction_hash", txHashes, &transactions)
	if err != nil {
		return transactions, fmt.Errorf("get transaction by hash: %w", err)
	}

	return transactions, nil
}
