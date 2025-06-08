package ethereum_test

import (
	"context"
	"errors"
	"fethcher/internal/ethereum"
	"fethcher/internal/ethereum/fake"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("EthService", func() {
	var (
		service    *ethereum.EthService
		fakeClient *fake.EthClient
		ctx        context.Context
		testErr    error
	)

	BeforeEach(func() {
		fakeClient = new(fake.EthClient)
		testErr = errors.New("test error")
		ctx = context.Background()
		service = ethereum.NewEthService(fakeClient)
	})

	Describe("FetchTransactions", func() {
		var (
			hashes    []string
			results   []*ethereum.Transaction
			err       error
			signedTx1 *types.Transaction
			signedTx2 *types.Transaction
			chainID   *big.Int
			tx1       *types.Transaction
			tx2       *types.Transaction
		)

		BeforeEach(func() {
			privateKey, err := crypto.GenerateKey()
			Expect(err).NotTo(HaveOccurred())

			chainID = big.NewInt(5)
			signer := types.LatestSignerForChainID(chainID)

			tx1 = types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), nil)
			tx2 = types.NewTransaction(1, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)

			signedTx1, _ = types.SignTx(tx1, signer, privateKey)
			signedTx2, _ = types.SignTx(tx2, signer, privateKey)

			hashes = []string{
				signedTx1.Hash().Hex(),
				signedTx2.Hash().Hex(),
			}

			fakeClient.NetworkIDReturns(chainID, nil)

			fakeClient.TransactionReceiptReturnsOnCall(0, &types.Receipt{
				Status:      1,
				BlockHash:   common.HexToHash("0xabc"),
				BlockNumber: big.NewInt(100),
			}, nil)
			fakeClient.TransactionReceiptReturnsOnCall(1, &types.Receipt{
				Status:      1,
				BlockHash:   common.HexToHash("0xdef"),
				BlockNumber: big.NewInt(101),
			}, nil)
		})

		JustBeforeEach(func() {
			results, err = service.FetchTransactions(ctx, hashes)
		})

		When("all transactions are fetched successfully", func() {
			BeforeEach(func() {
				fakeClient.TransactionByHashReturnsOnCall(0, signedTx1, false, nil)
				fakeClient.TransactionByHashReturnsOnCall(1, signedTx2, false, nil)
			})

			It("should return all transactions", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(results).To(HaveLen(2))
				Expect(results[0].TransactionHash).To(Or(Equal(signedTx1.Hash().Hex()), Equal(signedTx2.Hash().Hex())))
				Expect(results[1].TransactionHash).To(Or(Equal(signedTx1.Hash().Hex()), Equal(signedTx2.Hash().Hex())))

				Expect(fakeClient.TransactionByHashCallCount()).To(Equal(2))

				_, argHash1 := fakeClient.TransactionByHashArgsForCall(0)
				Expect(argHash1.Hex()).To(Or(Equal(signedTx1.Hash().Hex()), Equal(signedTx2.Hash().Hex())))

				_, argHash2 := fakeClient.TransactionByHashArgsForCall(1)
				Expect(argHash2.Hex()).To(Or(Equal(signedTx1.Hash().Hex()), Equal(signedTx2.Hash().Hex())))

				Expect(fakeClient.TransactionReceiptCallCount()).To(Equal(2))

				_, argRecHash1 := fakeClient.TransactionReceiptArgsForCall(0)
				Expect(argRecHash1).To(Or(Equal(signedTx1.Hash()), Equal(signedTx2.Hash())))

				_, argRecHash2 := fakeClient.TransactionReceiptArgsForCall(1)
				Expect(argRecHash2).To(Or(Equal(signedTx1.Hash()), Equal(signedTx2.Hash())))
			})
		})

		When("some transactions fail to fetch", func() {
			BeforeEach(func() {
				fakeClient.TransactionByHashReturnsOnCall(0, nil, false, testErr)
				fakeClient.TransactionByHashReturnsOnCall(1, signedTx2, false, nil)
			})

			It("should return partial results with error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("fetching transaction %q: %s", hashes[0], testErr.Error())))
				Expect(results).To(HaveLen(1))
				Expect(results[0].TransactionHash).To(Equal(signedTx2.Hash().Hex()))
			})
		})

		When("context is cancelled", func() {
			BeforeEach(func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				fakeClient.TransactionByHashStub = func(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
					select {
					case <-ctx.Done():
						return nil, false, ctx.Err()
					case <-time.After(100 * time.Millisecond):
						return tx1, false, nil
					}
				}
			})

			It("should return context cancelled error", func() {
				Expect(err).To(MatchError(context.Canceled))
			})
		})
	})

})
