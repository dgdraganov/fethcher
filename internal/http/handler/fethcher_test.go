package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"fethcher/internal/core"
	"fethcher/internal/http/handler"
	"fethcher/internal/http/handler/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

var _ = Describe("FethHandler", func() {
	var (
		fh            *handler.FethHandler
		fakeService   *fake.TransactionService
		fakeValidator *fake.RequestValidator
		fakeLogger    *zap.SugaredLogger
		w             *httptest.ResponseRecorder
		req           *http.Request
		testToken     string
		fakeErr       error
	)

	BeforeEach(func() {
		testToken = "test-token"
		fakeErr = errors.New("fake-error")
		fakeLogger = zap.NewNop().Sugar()
		fakeService = new(fake.TransactionService)
		fakeService.AuthenticateReturns(testToken, nil)
		fakeService.SaveUserTransactionsHistoryReturns(nil)
		fakeValidator = new(fake.RequestValidator)

		w = httptest.NewRecorder()
		fh = handler.NewFethHandler(fakeLogger, fakeValidator, fakeService)
	})

	Describe("HandleAuthenticate", func() {
		var (
			err      error
			response map[string]string
		)

		BeforeEach(func() {
			body := strings.NewReader(`{"username":"test","password":"pass"}`)
			req = httptest.NewRequest("POST", "/lime/authenticate", body)
			req.Header.Set("Content-Type", "application/json")

			fakeValidator.DecodeJSONPayloadStub = func(rec *http.Request, jsonPayload any) error {
				return json.NewDecoder(rec.Body).Decode(jsonPayload)
			}
		})

		JustBeforeEach(func() {
			fh.HandleAuthenticate(w, req)
		})

		When("authentication succeeds", func() {
			It("should return a token", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				decErr := json.NewDecoder(w.Body).Decode(&response)
				Expect(decErr).NotTo(HaveOccurred())
				Expect(response["token"]).To(Equal(testToken))
				Expect(fakeService.AuthenticateCallCount()).To(Equal(1))
				Expect(fakeValidator.DecodeJSONPayloadCallCount()).To(Equal(1))
				argReq, _ := fakeValidator.DecodeJSONPayloadArgsForCall(0)
				Expect(argReq).To(Equal(req))
			})
		})

		When("payload validation fails", func() {
			BeforeEach(func() {
				fakeValidator.DecodeJSONPayloadReturns(fakeErr)
			})

			It("should return status 400", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				Expect(w.Body.String()).To(ContainSubstring(fakeErr.Error()))
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeService.AuthenticateCallCount()).To(Equal(0))
				Expect(fakeValidator.DecodeJSONPayloadCallCount()).To(Equal(1))
			})
		})

		When("authentication fails due to incorrect credentials", func() {
			BeforeEach(func() {
				fakeService.AuthenticateReturns("", core.ErrIncorrectPassword)
			})

			It("should return 401 Unauthorized", func() {
				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				Expect(fakeService.AuthenticateCallCount()).To(Equal(1))
				Expect(fakeValidator.DecodeJSONPayloadCallCount()).To(Equal(1))
			})
		})
	})

	Describe("HandleGetTransactions", func() {
		var (
			req *http.Request
		)

		BeforeEach(func() {
			req = httptest.NewRequest("GET", "/lime/eth?transactionHashes=0x1&transactionHashes=0x2", nil)
		})

		JustBeforeEach(func() {
			fh.HandleGetTransactions(w, req)
		})

		When("transactions are fetched successfully", func() {
			BeforeEach(func() {
				fakeService.GetTransactionsReturns([]core.TransactionRecord{
					{TransactionHash: "0x1"},
					{TransactionHash: "0x2"},
				}, nil)
			})

			It("should return the transactions", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				var response map[string][]core.TransactionRecord
				err := json.NewDecoder(w.Body).Decode(&response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["transactions"]).To(HaveLen(2))
			})
			When("auth token is provided", func() {
				BeforeEach(func() {
					req.Header.Set("AUTH_TOKEN", testToken)
				})
				It("should save to user history", func() {
					Eventually(func() int {
						return fakeService.SaveUserTransactionsHistoryCallCount()
					}).Should(Equal(1))
				})
			})

		})

		When("query parameters are invalid", func() {
			BeforeEach(func() {
				req = httptest.NewRequest("GET", "/lime/eth?invalid=param", nil)
			})

			It("should return 400 Bad Request", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				Expect(fakeService.SaveUserTransactionsHistoryCallCount()).To(Equal(0))
				Expect(fakeService.GetTransactionsCallCount()).To(Equal(0))
			})
		})

		When("transaction service fails", func() {
			BeforeEach(func() {
				fakeService.GetTransactionsReturns(nil, fakeErr)
			})

			It("should return 500 Internal Server Error", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				Expect(w.Body.String()).To(ContainSubstring(fakeErr.Error()))
			})
		})
	})

	Describe("HandleGetTransactionsRLP", func() {
		var rlp string

		BeforeEach(func() {
			rlp = "rlp123"
			req = httptest.NewRequest("GET", "/lime/eth/"+rlp, nil)
		})

		JustBeforeEach(func() {
			fh.HandleGetTransactionsRLP(w, req)
		})

		When("rlp param is empty", func() {
			BeforeEach(func() {
				rlp = ""
				req = httptest.NewRequest("GET", "/lime/eth/", nil)
			})

			It("should return 400 Bad Request", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				Expect(w.Body.String()).To(ContainSubstring("rlp parameter is required"))
				Expect(fakeService.ParseRLPCallCount()).To(Equal(0))
				Expect(fakeService.GetTransactionsCallCount()).To(Equal(0))
			})
		})

		When("ParseRLP fails", func() {
			BeforeEach(func() {
				fakeService.ParseRLPReturns(nil, fakeErr)
			})

			It("should return 500 Internal Server Error", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				Expect(w.Body.String()).To(ContainSubstring(fakeErr.Error()))
			})
		})

		When("getting transactions fails", func() {
			BeforeEach(func() {
				fakeService.ParseRLPReturns([]string{"0x12"}, nil)
				fakeService.GetTransactionsReturns(nil, fakeErr)
			})

			It("should return 500 Internal Server Error", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				Expect(w.Body.String()).To(ContainSubstring(fakeErr.Error()))
			})
		})

		When("rlp is valid", func() {
			BeforeEach(func() {
				fakeService.ParseRLPReturns([]string{"0x12"}, nil)
				fakeService.GetTransactionsReturns([]core.TransactionRecord{
					{TransactionHash: "0x12"},
				}, nil)
			})

			It("should return 200 OK and transactions", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("0x12"))
				Expect(fakeService.ParseRLPCallCount()).To(Equal(1))
				Expect(fakeService.GetTransactionsCallCount()).To(Equal(1))
			})
		})
	})

	Describe("HandleGetAll", func() {
		JustBeforeEach(func() {
			fh.HandleGetAllTransactions(w, req)
		})

		When("GetAllDBTransactions succeeds", func() {
			BeforeEach(func() {
				req = httptest.NewRequest("GET", "/lime/all", nil)
				fakeService.GetAllDBTransactionsReturns([]core.TransactionRecord{
					{TransactionHash: "0x12"},
				}, nil)
			})

			It("should return 200 OK and transactions", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("0x12"))
				Expect(fakeService.GetAllDBTransactionsCallCount()).To(Equal(1))
			})
		})

		When("GetAllDBTransactions fails", func() {
			BeforeEach(func() {
				req = httptest.NewRequest("GET", "/lime/all", nil)
				fakeService.GetAllDBTransactionsReturns(nil, fakeErr)
			})

			It("should return 500 Internal Server Error", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				Expect(w.Body.String()).To(ContainSubstring(fakeErr.Error()))
			})
		})
	})

	Describe("HandleGetMy", func() {
		BeforeEach(func() {
			req = httptest.NewRequest("GET", "/lime/my", nil)
			req.Header.Set("AUTH_TOKEN", testToken)
		})
		JustBeforeEach(func() {
			fh.HandleGetMyTransactions(w, req)
		})

		When("GetUserTransactionsHistory succeeds", func() {
			BeforeEach(func() {
				fakeService.GetUserTransactionsHistoryReturns([]core.TransactionRecord{
					{TransactionHash: "0xuser"},
				}, nil)
			})

			It("should return 200 OK and user transactions", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("0xuser"))
			})
		})

		When("GetUserTransactionsHistory fails", func() {
			BeforeEach(func() {
				fakeService.GetUserTransactionsHistoryReturns(nil, fakeErr)
			})

			It("should return 500 Internal Server Error", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				Expect(w.Body.String()).To(ContainSubstring(fakeErr.Error()))
			})
		})

		When("no auth token is provided", func() {
			BeforeEach(func() {
				req.Header.Set("AUTH_TOKEN", "")
			})

			It("should return 401 Unauthorized", func() {
				Expect(w.Code).To(Equal(http.StatusUnauthorized))
				Expect(w.Body.String()).To(ContainSubstring("AUTH_TOKEN header is required"))
			})
		})
	})
})
