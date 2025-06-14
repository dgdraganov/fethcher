package repository_test

import (
	"context"
	"errors"
	"fethcher/internal/db"
	"fethcher/internal/repository"
	"fethcher/internal/repository/fake"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransactionRepository", func() {
	var (
		repo        *repository.TransactionRepository
		fakeStorage *fake.Storage
		ctx         context.Context
		fakeErr     error
	)

	BeforeEach(func() {
		fakeStorage = new(fake.Storage)
		repo = repository.NewTransactionRepository(fakeStorage)
		fakeErr = errors.New("fake error")
		ctx = context.Background()
	})

	Describe("Migrate", func() {
		var err error

		JustBeforeEach(func() {
			err = repo.MigrateTables(&repository.Transaction{}, &repository.User{}, &repository.UserTransaction{})
		})

		When("migration succeeds", func() {
			BeforeEach(func() {
				fakeStorage.MigrateTableReturns(nil)
			})

			It("should migrate tables", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeStorage.MigrateTableCallCount()).To(Equal(1))
				tables := fakeStorage.MigrateTableArgsForCall(0)
				Expect(tables).To(HaveLen(3))
				Expect(tables[0]).To(BeAssignableToTypeOf(&repository.Transaction{}))
				Expect(tables[1]).To(BeAssignableToTypeOf(&repository.User{}))
				Expect(tables[2]).To(BeAssignableToTypeOf(&repository.UserTransaction{}))
			})
		})

		When("migration fails", func() {
			BeforeEach(func() {
				fakeStorage.MigrateTableReturns(errors.New("migration error"))
			})

			It("should return an error", func() {
				Expect(err).To(MatchError("migrate table(s): migration error"))
			})
		})
	})

	Describe("SeedUserTable", func() {
		var err error

		JustBeforeEach(func() {
			err = repo.SeedUserTable(context.Background())
		})

		When("seeding succeeds", func() {
			BeforeEach(func() {
				fakeStorage.SeedTableReturns(nil)
			})

			It("should seed user table", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeStorage.SeedTableCallCount()).To(Equal(1))
				_, argUsers := fakeStorage.SeedTableArgsForCall(0)
				usersPtr, ok := argUsers.(*[]repository.User)
				Expect(ok).To(BeTrue())
				Expect(usersPtr).NotTo(BeNil())
				users := *usersPtr
				Expect(users).NotTo(BeNil())
				Expect(users[0].Username).To(Equal("alice"))
				Expect(users[1].Username).To(Equal("bob"))
				Expect(users[2].Username).To(Equal("carol"))
				Expect(users[3].Username).To(Equal("dave"))
			})
		})

		When("seeding fails", func() {
			BeforeEach(func() {
				fakeStorage.SeedTableReturns(fakeErr)
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(fakeErr))
				Expect(fakeStorage.SeedTableCallCount()).To(Equal(1))
			})
		})
	})

	Describe("SaveTransactions", func() {
		var (
			transactions []repository.Transaction
			err          error
		)

		BeforeEach(func() {
			transactions = []repository.Transaction{
				{
					TransactionHash: "0x123",
					BlockNumber:     100,
					From:            "0x007",
				},
			}
		})

		JustBeforeEach(func() {
			err = repo.SaveTransactions(ctx, transactions)
		})

		When("save transactions succeeds", func() {
			BeforeEach(func() {
				fakeStorage.InsertToTableReturns(nil)
			})

			It("should save transactions", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeStorage.InsertToTableCallCount()).To(Equal(1))
				_, arg := fakeStorage.InsertToTableArgsForCall(0)
				Expect(arg).To(Equal(&transactions))
			})
		})

		When("save transactions fails", func() {
			BeforeEach(func() {
				fakeStorage.InsertToTableReturns(fakeErr)
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(fakeErr))
			})
		})
	})

	Describe("GetUserHistory", func() {
		var (
			userID string
			err    error
			hashes []string
		)

		BeforeEach(func() {
			userID = uuid.NewString()
			fakeErr = errors.New("fake error")
		})

		JustBeforeEach(func() {
			hashes, err = repo.GetUserHistory(ctx, userID)
		})

		When("user has history", func() {
			BeforeEach(func() {
				fakeStorage.GetAllByStub = func(ctx context.Context, column string, value any, dest any) error {
					userTxs := dest.(*[]repository.UserTransaction)
					*userTxs = []repository.UserTransaction{
						{TransactionHash: "0x1"},
						{TransactionHash: "0x2"},
					}
					return nil
				}
			})

			It("should return transaction hashes", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(hashes).To(Equal([]string{"0x1", "0x2"}))

				Expect(fakeStorage.GetAllByCallCount()).To(Equal(1))
				_, col, val, usrTxs := fakeStorage.GetAllByArgsForCall(0)
				Expect(col).To(Equal("user_id"))
				Expect(val).To(Equal(userID))
				Expect(usrTxs).To(BeAssignableToTypeOf(&[]repository.UserTransaction{}))
			})
		})

		When("user not found", func() {
			BeforeEach(func() {
				fakeStorage.GetAllByReturns(db.ErrNotFound)
			})

			It("should return user not found error", func() {
				Expect(err).To(MatchError(repository.ErrUserNotFound))
			})
		})

		When("database error occurs", func() {
			BeforeEach(func() {
				fakeStorage.GetAllByReturns(fakeErr)
			})

			It("should return the error", func() {
				Expect(err).To(MatchError(fakeErr))
			})
		})
	})

	Describe("SaveUserHistory", func() {
		var (
			userID       string
			transactions []string
			err          error
		)

		BeforeEach(func() {
			userID = uuid.NewString()
			transactions = []string{"0x1", "0x2"}
		})

		JustBeforeEach(func() {
			err = repo.SaveUserHistory(ctx, userID, transactions)
		})

		When("save succeeds", func() {
			BeforeEach(func() {
				fakeStorage.InsertToTableReturns(nil)
			})

			It("should save user transactions", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeStorage.InsertToTableCallCount()).To(Equal(1))
				_, arg := fakeStorage.InsertToTableArgsForCall(0)
				Expect(arg).To(BeAssignableToTypeOf(&[]repository.UserTransaction{}))
			})
		})

		When("save fails", func() {
			BeforeEach(func() {
				fakeStorage.InsertToTableReturns(fakeErr)
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(fakeErr))
			})
		})

		When("transactions are empty", func() {
			BeforeEach(func() {
				transactions = []string{}
			})
			It("should return immediately", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeStorage.InsertToTableCallCount()).To(Equal(0))
			})
		})
	})

	Describe("GetUserFromDB", func() {
		var (
			user     repository.User
			err      error
			username string
			testUser repository.User
		)

		BeforeEach(func() {
			username = "alice"
			testUser = repository.User{
				ID:           uuid.NewString(),
				Username:     username,
				PasswordHash: "hashed_password",
			}
		})
		JustBeforeEach(func() {
			user, err = repo.GetUserFromDB(ctx, username)
		})

		When("user exists", func() {
			BeforeEach(func() {
				fakeStorage.GetOneByStub = func(ctx context.Context, column string, value any, dest any) error {
					user := dest.(*repository.User)
					*user = testUser
					return nil
				}
			})

			It("should return the user", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(user.Username).To(Equal(username))

				Expect(fakeStorage.GetOneByCallCount()).To(Equal(1))
				_, col, val, _ := fakeStorage.GetOneByArgsForCall(0)
				Expect(col).To(Equal("username"))
				Expect(val).To(Equal(username))
			})
		})

		When("user doesn't exist", func() {
			BeforeEach(func() {
				fakeStorage.GetOneByReturns(repository.ErrUserNotFound)
			})

			It("should return user not found error", func() {
				Expect(err).To(MatchError(repository.ErrUserNotFound))
			})
		})

		When("database error occurs", func() {
			BeforeEach(func() {
				fakeStorage.GetOneByReturns(fakeErr)
			})

			It("should return the error", func() {
				Expect(err).To(MatchError(fakeErr))
			})
		})
	})

	Describe("GetTransactionsByHash", func() {
		var (
			txHashes     []string
			transactions []repository.Transaction
			err          error
		)

		BeforeEach(func() {
			txHashes = []string{"0x1", "0x2"}
		})
		JustBeforeEach(func() {
			transactions, err = repo.GetTransactionsByHash(ctx, txHashes)
		})

		When("transactions exist", func() {
			BeforeEach(func() {
				fakeStorage.GetAllByStub = func(ctx context.Context, column string, value any, dest any) error {
					txs := dest.(*[]repository.Transaction)
					*txs = []repository.Transaction{
						{TransactionHash: "0x1"},
						{TransactionHash: "0x2"},
					}
					return nil
				}
			})

			It("should return transactions", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(2))

				Expect(fakeStorage.GetAllByCallCount()).To(Equal(1))
				_, col, val, _ := fakeStorage.GetAllByArgsForCall(0)
				Expect(col).To(Equal("transaction_hash"))
				Expect(val).To(Equal(txHashes))
			})
		})

		When("no transactions exist", func() {
			BeforeEach(func() {
				fakeStorage.GetAllByReturns(nil)
			})

			It("should return empty slice", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(BeEmpty())
			})
		})

		When("database error occurs", func() {
			BeforeEach(func() {
				fakeStorage.GetAllByReturns(fakeErr)
			})

			It("should return the error", func() {
				Expect(err).To(MatchError(fakeErr))
			})
		})
	})
})
