package core

import (
	"context"
	"encoding/hex"
	"errors"
	"fethcher/internal/repository"
	tokenIssuer "fethcher/pkg/jwt"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var ErrIncorrectPassword error = errors.New("incorrect password")
var ErrUserNotFound error = errors.New("user not found")

// Fethcher is a struct that provides methods to interact with the Ethereum node and the database.
type Fethcher struct {
	logs       *zap.SugaredLogger
	repo       Repository
	jwtIssuer  JWTIssuer
	ethService EthereumService
}

// NewFethcher is a constructor function for the Fethcher type.
func NewFethcher(logger *zap.SugaredLogger, repo Repository, jwt JWTIssuer, ethereumService EthereumService) *Fethcher {
	return &Fethcher{
		logs:       logger,
		repo:       repo,
		jwtIssuer:  jwt,
		ethService: ethereumService,
	}
}

// Authenticate checks the provided username and password against the database. If the credentials are valid, it generates a JWT token for the user.
func (f *Fethcher) Authenticate(ctx context.Context, msg AuthMessage) (string, error) {
	user, err := f.repo.GetUserFromDB(ctx, msg.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
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

// GetTransactions retrieves transactions by their hashes. It first checks the database for existing transactions,
func (f *Fethcher) GetTransactions(ctx context.Context, transactionsHashes []string) ([]TransactionRecord, error) {

	records := make([]TransactionRecord, 0, len(transactionsHashes))
	dbTxs, err := f.getTransactionsFromDB(ctx, transactionsHashes)
	if err != nil {
		return nil, fmt.Errorf("get transactions from db: %w", err)
	}

	f.logs.Infow("transactions fetched from db", "count", len(dbTxs))

	records = append(records, dbTxs...)

	recordsMap := make(map[string]struct{})
	for _, rec := range records {
		recordsMap[rec.TransactionHash] = struct{}{}
	}

	if len(records) == len(transactionsHashes) {
		f.logs.Infow("all transactions found in DB", "count", len(records))
		return records, nil
	}

	missingTransactions := make([]string, 0, len(transactionsHashes)-len(records))

	for _, transactionHash := range transactionsHashes {
		if _, ok := recordsMap[transactionHash]; !ok {
			missingTransactions = append(missingTransactions, transactionHash)
		}
	}

	nodeTxs, err := f.getTransactionsFromNode(ctx, missingTransactions)
	if err != nil {
		f.logs.Errorw("getting transactions from node", "error", err)
	}

	records = append(records, nodeTxs...)

	f.logs.Infow("caching transactions from eth node to DB", "transactions", nodeTxs)

	go func() {
		err = f.saveTransactionsToDB(ctx, nodeTxs)
		if err != nil {
			f.logs.Errorw("failed to save transactions to DB", "error", err, "count", len(nodeTxs))
		}
	}()

	return records, nil
}

// SaveUserTransactionsHistory saves the transaction history for a user based on the provided JWT token and transaction hashes.
func (f *Fethcher) SaveUserTransactionsHistory(ctx context.Context, token string, transactionsHashes []string) error {
	if len(transactionsHashes) == 0 {
		return nil
	}

	claims, err := f.jwtIssuer.Validate(token)
	if err != nil {
		return fmt.Errorf("validate jwt token: %w", err)
	}

	userId := claims["sub"].(string)

	err = f.repo.SaveUserHistory(ctx, userId, transactionsHashes)
	if err != nil {
		return fmt.Errorf("save user history: %w", err)
	}

	f.logs.Infow("user history saved", "userId", userId, "transactions", transactionsHashes)
	return nil
}

// GetUserTransactionsHistory retrieves the transaction history for a user based on the provided JWT token.
func (f *Fethcher) GetUserTransactionsHistory(ctx context.Context, token string) ([]TransactionRecord, error) {
	claims, err := f.jwtIssuer.Validate(token)
	if err != nil {
		return []TransactionRecord{}, fmt.Errorf("validate jwt token: %w", err)
	}

	userId := claims["sub"].(string)

	f.logs.Infow("getting user transactions history", "userId", userId)

	transactionsHashes, err := f.repo.GetUserHistory(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("get user history: %w", err)
	}

	txRecords, err := f.getTransactionsFromDB(ctx, transactionsHashes)
	if err != nil {
		return nil, fmt.Errorf("get transactions by hash: %w", err)
	}

	f.logs.Infow("user transactions history fetched from DB", "userId", userId, "transactionsCount", len(txRecords))

	return txRecords, nil
}

// GetAllDBTransactions retrieves all transactions from the database and returns them as a slice of TransactionRecord.
func (f *Fethcher) GetAllDBTransactions(ctx context.Context) ([]TransactionRecord, error) {
	transactions, err := f.repo.GetAllTransactions(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all transactions: %w", err)
	}
	records := f.repoTransactionToRecord(transactions)

	return records, nil
}

// ParseRLP decodes a hex-encoded RLP string into a slice of transaction hashes.
func (f *Fethcher) ParseRLP(rlphex string) ([]string, error) {
	data, err := hex.DecodeString(rlphex)
	if err != nil {
		return nil, fmt.Errorf("decode hex string: %w", err)
	}

	var txHashBytes [][]byte
	if err := rlp.DecodeBytes(data, &txHashBytes); err != nil {
		return nil, fmt.Errorf("decode rlp bytes: %w", err)
	}

	txHashes := make([]string, len(txHashBytes))
	for i, b := range txHashBytes {
		txHashes[i] = fmt.Sprintf("0x%s", hex.EncodeToString(b))
	}
	return txHashes, nil
}

func (f *Fethcher) saveTransactionsToDB(ctx context.Context, transactionRecords []TransactionRecord) error {
	transactions := make([]repository.Transaction, 0, len(transactionRecords))
	for _, tx := range transactionRecords {
		transactions = append(transactions, repository.Transaction{
			TransactionHash:   tx.TransactionHash,
			TransactionStatus: tx.TransactionStatus,
			BlockHash:         tx.BlockHash,
			BlockNumber:       tx.BlockNumber,
			From:              tx.From,
			To:                tx.To,
			ContractAddress:   tx.ContractAddress,
			LogsCount:         tx.LogsCount,
			Input:             tx.Input,
			Value:             tx.Value,
		})
	}

	if err := f.repo.SaveTransactions(ctx, transactions); err != nil {
		return fmt.Errorf("repo save transactions: %w", err)
	}
	return nil
}

func (f *Fethcher) getTransactionsFromDB(ctx context.Context, transactionsHashes []string) ([]TransactionRecord, error) {

	dbTransactions, err := f.repo.GetTransactionsByHash(ctx, transactionsHashes)
	if err != nil {
		return nil, fmt.Errorf("get transactions by hash: %w", err)
	}

	records := f.repoTransactionToRecord(dbTransactions)
	return records, nil
}

func (f *Fethcher) repoTransactionToRecord(transactions []repository.Transaction) []TransactionRecord {
	records := make([]TransactionRecord, len(transactions))
	for i, tx := range transactions {
		records[i] = TransactionRecord{
			TransactionHash:   tx.TransactionHash,
			TransactionStatus: tx.TransactionStatus,
			BlockHash:         tx.BlockHash,
			BlockNumber:       tx.BlockNumber,
			From:              tx.From,
			To:                tx.To,
			ContractAddress:   tx.ContractAddress,
			LogsCount:         tx.LogsCount,
			Input:             tx.Input,
			Value:             tx.Value,
		}
	}
	return records
}

func (f *Fethcher) getTransactionsFromNode(ctx context.Context, transactionsHashes []string) ([]TransactionRecord, error) {

	transactions, err := f.ethService.FetchTransactions(ctx, transactionsHashes)

	records := make([]TransactionRecord, len(transactions))
	for i, tx := range transactions {
		records[i] = TransactionRecord{
			TransactionHash:   tx.TransactionHash,
			TransactionStatus: tx.TransactionStatus,
			BlockHash:         tx.BlockHash,
			BlockNumber:       tx.BlockNumber,
			From:              tx.From,
			To:                tx.To,
			ContractAddress:   tx.ContractAddress,
			LogsCount:         tx.LogsCount,
			Input:             tx.Input,
			Value:             tx.Value,
		}
	}

	return records, err
}
