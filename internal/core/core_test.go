package core_test

import (
	"context"
	"errors"
	"fethcher/internal/core"
	"fethcher/internal/core/fake"
	"fethcher/internal/ethereum"
	"fethcher/internal/repository"
	tokenIssuer "fethcher/pkg/jwt"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("Fethcher", func() {
	var (
		fakeRepo   *fake.Repository
		fakeJWT    *fake.JWTIssuer
		fakeEth    *fake.EthereumService
		fakeLogger *zap.SugaredLogger
		ctx        context.Context

		fetcher *core.Fethcher

		fakeErr error
	)

	BeforeEach(func() {
		fakeRepo = new(fake.Repository)
		fakeJWT = new(fake.JWTIssuer)
		fakeEth = new(fake.EthereumService)
		fakeLogger = zap.NewNop().Sugar()
		ctx = context.Background()

		fetcher = core.NewFethcher(fakeLogger, fakeRepo, fakeJWT, fakeEth)

		fakeErr = errors.New("fake error")
	})

	Describe("Authenticate", func() {
		var (
			authMsg        core.AuthMessage
			token          string
			err            error
			userId         string
			tokenInfo      tokenIssuer.TokenInfo
			hashedPassword string
			genToken       *jwt.Token
		)

		BeforeEach(func() {
			userId = uuid.New().String()
			hashedPassword = "$2a$10$1MZHKX./8Dxi9t.F1/gnx.njCcEty299Hx01GLEms2moa3brpT0ky" // bcrypt hash of "testpass"
			genToken = jwt.New(jwt.SigningMethodHS256)

			authMsg = core.AuthMessage{
				Username: "testuser",
				Password: "testpass",
			}

			tokenInfo = tokenIssuer.TokenInfo{
				UserName:   authMsg.Username,
				Subject:    userId,
				Expiration: 24,
			}
		})

		JustBeforeEach(func() {
			token, err = fetcher.Authenticate(ctx, authMsg)
		})

		When("user exists and password matches", func() {
			BeforeEach(func() {
				fakeRepo.GetUserFromDBReturns(repository.User{
					Username:     authMsg.Username,
					PasswordHash: hashedPassword,
					ID:           userId,
				}, nil)

				fakeJWT.GenerateReturns(genToken)
				fakeJWT.SignReturns("signed.token", nil)

			})

			It("should return a signed token", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(token).To(Equal("signed.token"))

				Expect(fakeRepo.GetUserFromDBCallCount()).To(Equal(1))
				_, username, password := fakeRepo.GetUserFromDBArgsForCall(0)
				Expect(username).To(Equal(authMsg.Username))
				Expect(password).To(Equal(authMsg.Password))

				Expect(fakeJWT.GenerateCallCount()).To(Equal(1))
				argGen := fakeJWT.GenerateArgsForCall(0)
				Expect(argGen).To(Equal(tokenInfo))

				Expect(fakeJWT.SignCallCount()).To(Equal(1))
				argSign := fakeJWT.SignArgsForCall(0)
				Expect(argSign).To(Equal(genToken))
			})
		})

		When("user does not exist", func() {
			BeforeEach(func() {
				fakeRepo.GetUserFromDBReturns(repository.User{}, repository.ErrUserNotFound)
			})

			It("should return user not found error", func() {
				Expect(err).To(MatchError(core.ErrUserNotFound))
			})
		})

		When("password does not match", func() {
			BeforeEach(func() {
				fakeRepo.GetUserFromDBReturns(repository.User{
					Username:     authMsg.Username,
					PasswordHash: hashedPassword, // bcrypt hash of "testpass"
				}, nil)
				authMsg.Password = "wrongpass"
			})

			It("should return incorrect password error", func() {
				Expect(err).To(MatchError(core.ErrIncorrectPassword))
			})
		})

		When("token signing fails", func() {
			BeforeEach(func() {
				fakeRepo.GetUserFromDBReturns(repository.User{
					Username:     authMsg.Username,
					PasswordHash: hashedPassword,
					ID:           userId,
				}, nil)
				fakeJWT.SignReturns("", fakeErr)
			})

			It("should return signing error", func() {
				Expect(err).To(MatchError(fakeErr))
			})
		})
	})

	Describe("GetTransactions", func() {
		var (
			txHashes  []string
			txRecords []core.TransactionRecord
			err       error
		)

		BeforeEach(func() {
			txHashes = []string{"0x1", "0x2"}
		})

		JustBeforeEach(func() {
			txRecords, err = fetcher.GetTransactions(ctx, txHashes)
		})

		When("transactions exist in DB", func() {
			BeforeEach(func() {
				fakeRepo.GetTransactionsByHashReturns([]repository.Transaction{
					{TransactionHash: "0x1"},
					{TransactionHash: "0x2"},
				}, nil)
			})

			It("should return transactions from DB", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(txRecords).To(HaveLen(2))
				Expect(fakeRepo.GetTransactionsByHashCallCount()).To(Equal(1))
				_, argTxs := fakeRepo.GetTransactionsByHashArgsForCall(0)
				Expect(argTxs).To(Equal(txHashes))
				Expect(fakeEth.FetchTransactionsCallCount()).To(Equal(0))
			})
		})

		When("one or more transactions missing from DB", func() {
			BeforeEach(func() {
				fakeRepo.GetTransactionsByHashReturns([]repository.Transaction{
					{TransactionHash: "0x1"},
				}, nil)
				fakeEth.FetchTransactionsReturns([]*ethereum.Transaction{
					{TransactionHash: "0x2"},
				}, nil)
			})

			It("fetches missing transactions from ethereum node", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(txRecords).To(HaveLen(2))
				Expect(fakeRepo.GetTransactionsByHashCallCount()).To(Equal(1))
				Expect(fakeEth.FetchTransactionsCallCount()).To(Equal(1))
				_, argTxs := fakeEth.FetchTransactionsArgsForCall(0)
				Expect(argTxs).To(Equal([]string{"0x2"}))
			})
		})

		When("getting txs from db fails", func() {
			BeforeEach(func() {
				fakeRepo.GetTransactionsByHashReturns(nil, fakeErr)
			})

			It("should return error", func() {
				Expect(err).To(MatchError(fakeErr))
			})
		})

		When("node fetch fails", func() {
			BeforeEach(func() {
				fakeRepo.GetTransactionsByHashReturns([]repository.Transaction{
					{TransactionHash: "0x1"},
				}, nil)
				fakeEth.FetchTransactionsReturns(nil, fakeErr)
			})

			It("should return partial results with error", func() {
				Expect(txRecords).To(HaveLen(1))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("ParseRLP", func() {
		var (
			transactions []string
			err          error
			rlphex       string
			parsedSlice  []string
		)

		BeforeEach(func() {
			rlphex = "f90110b842307866633262336236646233386135316462336239636239356465323962373139646538646562393936333036323665346234623939646630353666666237663265b842307834383630336637616466663766626663326131306232326136373130333331656536386632653464316364373361353834643537633838323164663739333536b842307863626339323065376262383963626362353430613436396131363232366266313035373832353238336162386561633366343564303038313165656638613634b842307836643630346666633634346132383266636138636238653737386531653366383234356438626431643439333236653330313661336338373862613063626264"
			parsedSlice = []string{
				"0x307866633262336236646233386135316462336239636239356465323962373139646538646562393936333036323665346234623939646630353666666237663265",
				"0x307834383630336637616466663766626663326131306232326136373130333331656536386632653464316364373361353834643537633838323164663739333536",
				"0x307863626339323065376262383963626362353430613436396131363232366266313035373832353238336162386561633366343564303038313165656638613634",
				"0x307836643630346666633634346132383266636138636238653737386531653366383234356438626431643439333236653330313661336338373862613063626264",
			}
		})

		JustBeforeEach(func() {
			transactions, err = fetcher.ParseRLP(rlphex)
		})

		When("RLP is valid", func() {
			It("returns the transaction hashes", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(Equal(parsedSlice))
			})
		})

		When("RLP is NOT valid", func() {
			BeforeEach(func() {
				rlphex = "f90110b842307866633262336236646233386135"
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(transactions).To(BeNil())
			})
		})
	})

	Describe("SaveUserTransactionsHistory", func() {
		var (
			token    string
			txHashes []string
			err      error
			userId   string
		)

		BeforeEach(func() {
			token = "valid.token"
			userId = "user123"
			txHashes = []string{"0x1", "0x2"}
		})

		JustBeforeEach(func() {
			err = fetcher.SaveUserTransactionsHistory(ctx, token, txHashes)
		})

		When("token is valid and save succeeds", func() {
			BeforeEach(func() {
				fakeJWT.ValidateReturns(jwt.MapClaims{"sub": userId}, nil)
				fakeRepo.SaveUserHistoryReturns(nil)
			})

			It("should save user history", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeJWT.ValidateCallCount()).To(Equal(1))
				argToken := fakeJWT.ValidateArgsForCall(0)
				Expect(argToken).To(Equal(token))
				Expect(fakeRepo.SaveUserHistoryCallCount()).To(Equal(1))
				_, argUserId, argTxHashes := fakeRepo.SaveUserHistoryArgsForCall(0)
				Expect(argUserId).To(Equal(userId))
				Expect(argTxHashes).To(Equal(txHashes))
			})
		})

		When("token is invalid", func() {
			BeforeEach(func() {
				fakeJWT.ValidateReturns(nil, fakeErr)
			})

			It("should return validation error", func() {
				Expect(err).To(MatchError(fakeErr))
			})
		})

		When("save fails", func() {
			BeforeEach(func() {
				fakeJWT.ValidateReturns(jwt.MapClaims{"sub": "user123"}, nil)
				fakeRepo.SaveUserHistoryReturns(fakeErr)
			})

			It("should return save error", func() {
				Expect(err).To(MatchError(fakeErr))
			})
		})

		When("transactions are empty", func() {
			BeforeEach(func() {
				txHashes = []string{}
			})

			It("should skip saving", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeRepo.SaveUserHistoryCallCount()).To(Equal(0))
			})
		})
	})

	Describe("GetUserTransactionsHistory", func() {
		var (
			token     string
			txRecords []core.TransactionRecord
			err       error
		)

		BeforeEach(func() {
			token = "valid.token"
		})

		JustBeforeEach(func() {
			txRecords, err = fetcher.GetUserTransactionsHistory(ctx, token)
		})

		When("user has transaction history", func() {
			BeforeEach(func() {
				fakeJWT.ValidateReturns(jwt.MapClaims{"sub": "user123"}, nil)
				fakeRepo.GetUserHistoryReturns([]string{"0x1", "0x2"}, nil)
				fakeRepo.GetTransactionsByHashReturns([]repository.Transaction{
					{TransactionHash: "0x1"},
					{TransactionHash: "0x2"},
				}, nil)
			})

			It("should return user transactions", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(txRecords).To(HaveLen(2))
				Expect(fakeJWT.ValidateCallCount()).To(Equal(1))
				Expect(fakeRepo.GetUserHistoryCallCount()).To(Equal(1))
				Expect(fakeRepo.GetTransactionsByHashCallCount()).To(Equal(1))
			})
		})

		When("user has no history", func() {
			BeforeEach(func() {
				fakeJWT.ValidateReturns(jwt.MapClaims{"sub": "user123"}, nil)
				fakeRepo.GetUserHistoryReturns(nil, repository.ErrUserNotFound)
			})

			It("should return not found error", func() {
				Expect(err).To(MatchError(repository.ErrUserNotFound))
			})
		})

		When("token is invalid", func() {
			BeforeEach(func() {
				fakeJWT.ValidateReturns(nil, fakeErr)
			})

			It("should return validation error", func() {
				Expect(err).To(MatchError(fakeErr))
				Expect(fakeJWT.ValidateCallCount()).To(Equal(1))
			})
		})
	})
})
