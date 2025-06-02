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

type AuthMessage struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Fethcher struct {
	logs       *zap.SugaredLogger
	repo       Repository
	jwtIssuer  JWTIssuer
	ethService EthereumService
}

func NewFethcher(logger *zap.SugaredLogger, repo Repository, jwt JWTIssuer, ethereumService EthereumService) *Fethcher {
	return &Fethcher{
		logs:       logger,
		repo:       repo,
		jwtIssuer:  jwt,
		ethService: ethereumService,
	}
}

func (f *Fethcher) Authenticate(msg AuthMessage) (string, error) {
	user, err := f.repo.GetUserFromDB(msg.Username, msg.Password)
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

	f.logs.Infow("transactions fetched from ethereum", "count", len(nodeTxs))

	records = append(records, nodeTxs...)

	f.logs.Infow("caching transactions to DB", "count", len(records))

	err = f.saveTransactionsToDB(ctx, records)
	if err != nil {
		return records, fmt.Errorf("save transactions to db: %w", err)
	}

	return records, nil
}

func (f *Fethcher) GetTransactionsRLP(ctx context.Context, rlphex string) ([]TransactionRecord, error) {
	transactionsHashes, err := f.parseRLP(rlphex)
	if err != nil {
		return nil, fmt.Errorf("parse rlp: %w", err)
	}

	transactions, err := f.GetTransactions(ctx, transactionsHashes)
	if err != nil {
		return nil, fmt.Errorf("get transactions by hash: %w", err)
	}

	return transactions, err
}

func (f *Fethcher) SaveUserTransactionsHistory(ctx context.Context, token string, transactionsHashes []string) error {
	if len(transactionsHashes) == 0 {
		return nil
	}

	claims, err := f.jwtIssuer.Validate(token)
	if err != nil {
		return fmt.Errorf("validate jwt token: %w", err)
	}

	userId := claims["sub"].(string)

	f.repo.SaveUserHistory(userId, transactionsHashes)
	if err != nil {
		return fmt.Errorf("save user history: %w", err)
	}

	f.logs.Infow("user history saved", "userId", userId, "transactionsCount", len(transactionsHashes))
	return nil
}

func (f *Fethcher) GetUserTransactionsHistory(ctx context.Context, token string) ([]TransactionRecord, error) {
	claims, err := f.jwtIssuer.Validate(token)
	if err != nil {
		return []TransactionRecord{}, fmt.Errorf("validate jwt token: %w", err)
	}

	userId := claims["sub"].(string)

	f.logs.Infow("getting user transactions history", "userId", userId)

	transactionsHashes, err := f.repo.GetUserHistory(userId)
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
	err := f.repo.SaveTransactions(transactions)
	if err != nil {
		return fmt.Errorf("repo save transactions: %w", err)
	}
	return nil
}

func (f *Fethcher) getTransactionsFromDB(ctx context.Context, transactionsHashes []string) ([]TransactionRecord, error) {

	dbTransactions, err := f.repo.GetTransactionsByHash(transactionsHashes)
	if err != nil {
		return nil, fmt.Errorf("get transactions by hash: %w", err)
	}

	records := make([]TransactionRecord, len(dbTransactions))
	for i, tx := range dbTransactions {
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
	return records, nil
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

func (f *Fethcher) parseRLP(rlphex string) ([]string, error) {
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
