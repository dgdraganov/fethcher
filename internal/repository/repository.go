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

func (r *TransactionRepository) MigrateTables(tables ...any) error {
	err := r.db.MigrateTable(tables...)
	if err != nil {
		return fmt.Errorf("migrate table(s): %w", err)
	}
	return err
}

func (r *TransactionRepository) SeedUserTable(ctx context.Context) error {

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
	err := r.db.SeedTable(ctx, &users)
	if err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	return nil
}

func (r *TransactionRepository) SaveTransactions(ctx context.Context, transactions []Transaction) error {
	err := r.db.InsertToTable(ctx, &transactions)
	if err != nil {
		return fmt.Errorf("save to table: %w", err)
	}

	return nil
}

// GetUserHistory receives userID and retrieves the user transctions history.
func (r *TransactionRepository) GetUserHistory(ctx context.Context, userID string) ([]string, error) {
	var userTransactions []UserTransaction

	err := r.db.GetAllBy(ctx, "user_id", userID, &userTransactions)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, fmt.Errorf("get all by user_is: %w", ErrUserNotFound)
		}
		return nil, fmt.Errorf("get all by user_is: %w", err)
	}

	txHashes := make([]string, 0, len(userTransactions))
	for _, tx := range userTransactions {
		txHashes = append(txHashes, tx.TransactionHash)
	}

	return txHashes, nil
}

// SaveUserHistory receives an userID and a slice of transaction hashes and saves the user query history in the DB
func (r *TransactionRepository) SaveUserHistory(ctx context.Context, userID string, transactions []string) error {
	if len(transactions) == 0 {
		return nil
	}

	var dbUserTransactions []UserTransaction
	err := r.db.GetAllBy(ctx, "user_id", userID, &dbUserTransactions)
	if err != nil {
		return fmt.Errorf("get user transactions from db: %w", err)
	}

	var dbUserTxsMap = make(map[string]struct{}, len(dbUserTransactions))
	for _, tx := range dbUserTransactions {
		dbUserTxsMap[tx.TransactionHash] = struct{}{}
	}

	userTransactions := []UserTransaction{}
	for _, tx := range transactions {
		if _, exists := dbUserTxsMap[tx]; !exists {
			userTransactions = append(userTransactions, UserTransaction{
				UserID:          userID,
				TransactionHash: tx,
			})
		}
	}

	err = r.db.InsertToTable(ctx, &userTransactions)
	if err != nil {
		return fmt.Errorf("insert into table user_transactions: %w", err)
	}

	return nil
}

// GetUserFromDB receives an username and retrieves the user data from the DB.
func (r *TransactionRepository) GetUserFromDB(ctx context.Context, username string) (User, error) {
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

// GetTransactionsByHash receives a slice of transaction hashes and tries to retrieve them from the DB. The transactoins that cannot be found in the DB are then queried for from the Ethereum network and cached in the DB.
func (r *TransactionRepository) GetTransactionsByHash(ctx context.Context, txHashes []string) ([]Transaction, error) {
	transactions := []Transaction{}
	err := r.db.GetAllBy(ctx, "transaction_hash", txHashes, &transactions)
	if err != nil {
		return transactions, fmt.Errorf("get transaction by hash: %w", err)
	}

	return transactions, nil
}

// GetAllTransactions retrieves all transactions from the DB
func (r *TransactionRepository) GetAllTransactions(ctx context.Context) ([]Transaction, error) {
	transactions := []Transaction{}
	err := r.db.GetAll(ctx, &transactions)
	if err != nil {
		return nil, fmt.Errorf("get all transactions from db: %w", err)
	}
	return transactions, nil
}
