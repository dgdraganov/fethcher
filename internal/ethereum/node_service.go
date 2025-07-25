package ethereum

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EthService defines the interface for interacting with Ethereum transactions.
type EthService struct {
	client EthClient
}

// NewEthService is a constructor function for the EthService type.
func NewEthService(ethClient EthClient) *EthService {
	return &EthService{
		client: ethClient,
	}
}

// FetchTransactions fetches multiple transactions by their hashes concurrently.
func (s *EthService) FetchTransactions(ctx context.Context, hashes []string) ([]*Transaction, error) {
	resultsChan := make(chan *TxResult)

	var waitGrp sync.WaitGroup
	for ind, hashStr := range hashes {
		waitGrp.Add(1)
		go func(i int, hashStr string) {
			defer waitGrp.Done()
			hash := common.HexToHash(hashStr)
			res := s.getTransactionByHash(ctx, hash)
			if res.Error != nil {
				res.Error = fmt.Errorf("fetching transaction %q: %w", hashStr, res.Error)
			}
			resultsChan <- res
		}(ind, hashStr)
	}

	go func() {
		waitGrp.Wait()
		close(resultsChan)
	}()

	var results []*Transaction

	var aggrErr error
	for result := range resultsChan {
		if result.Error != nil {
			aggrErr = errors.Join(aggrErr, result.Error)
			continue
		}

		results = append(results, result.Transaction)
	}

	return results, aggrErr
}

func (s *EthService) getTransactionByHash(ctx context.Context, hash common.Hash) *TxResult {
	tx, _, err := s.client.TransactionByHash(ctx, hash)
	if err != nil {
		return &TxResult{nil, err}
	}

	receipt, err := s.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return &TxResult{nil, err}
	}

	chainID, err := s.client.NetworkID(ctx)
	if err != nil {
		return &TxResult{nil, err}
	}

	signer := types.LatestSignerForChainID(chainID)

	from, err := types.Sender(signer, tx)
	if err != nil {
		return &TxResult{nil, err}
	}

	var to string
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	var contractAddress *string

	if receipt.ContractAddress != (common.Address{}) {
		addr := receipt.ContractAddress.Hex()
		contractAddress = &addr
	}

	return &TxResult{
		Transaction: &Transaction{
			TransactionHash:   tx.Hash().Hex(),
			TransactionStatus: receipt.Status,
			BlockHash:         receipt.BlockHash.Hex(),
			BlockNumber:       receipt.BlockNumber.Uint64(),
			From:              from.Hex(),
			To:                toPtr(to),
			ContractAddress:   contractAddress,
			LogsCount:         len(receipt.Logs),
			Input:             fmt.Sprintf("0x%x", tx.Data()),
			Value:             tx.Value().String(),
		},
		Error: nil,
	}
}

func toPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
