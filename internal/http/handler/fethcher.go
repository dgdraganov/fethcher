package handler

import (
	"encoding/json"
	"errors"
	"fethcher/internal/core"
	"fethcher/internal/http/handler/middleware"
	"fethcher/internal/http/payload"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

var (
	Authenticate       = "POST /lime/authenticate"
	GetTransactions    = "GET /lime/eth"
	GetTransactionsRLP = "GET /lime/eth/{rlpHash}"
	GetAllTransactions = "GET /lime/all"
	GetMyTransactions  = "GET /lime/my"
)

type FethHandler struct {
	logs             *zap.SugaredLogger
	requestValidator RequestValidator
	fethcher         TransactionService
}

func NewFethHandler(logger *zap.SugaredLogger, requestValidator RequestValidator, transactionService TransactionService) *FethHandler {
	return &FethHandler{
		logs:             logger,
		requestValidator: requestValidator,
		fethcher:         transactionService,
	}
}

func (h *FethHandler) HandleAuthenticate(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	reqIdCtx := r.Context().Value(middleware.RequestIDKey)
	if reqIdCtx != nil {
		requestId = reqIdCtx.(string)
	}

	var payload payload.AuthRequest
	err := h.requestValidator.DecodeJSONPayload(r, &payload)
	if err != nil || payload.Validate() != nil {
		h.respond(w, Response{
			Message: "Could not authenticate",
			Error:   fmt.Errorf("invalid request payload: %w", err).Error(),
		}, http.StatusBadRequest,
			requestId)
		h.logs.Errorw("failed to decode and validate request payload",
			"error", err,
			"handler", Authenticate,
			"request_id", requestId)
		return
	}

	token, err := h.fethcher.Authenticate(r.Context(), payload.ToMessage())
	if err != nil {
		resp := Response{
			Message: "Login failed",
		}
		httpCode := http.StatusInternalServerError
		if errors.Is(err, core.ErrUserNotFound) {
			httpCode = http.StatusUnauthorized
			resp.Error = err.Error()
		} else if errors.Is(err, core.ErrIncorrectPassword) {
			httpCode = http.StatusUnauthorized
			resp.Error = err.Error()
		} else {
			httpCode = http.StatusInternalServerError
			resp.Error = "unexpected error occurred"
		}

		h.respond(w, resp, httpCode, requestId)
		h.logs.Errorw("authentication failed",
			"error", err,
			"handler", Authenticate,
			"request_id", requestId)
		return
	}

	resp := map[string]string{
		"token": token,
	}
	h.respond(w, resp, http.StatusOK, requestId)
}

func (h *FethHandler) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	reqIdCtx := r.Context().Value(middleware.RequestIDKey)
	if reqIdCtx != nil {
		requestId = reqIdCtx.(string)
	}

	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		h.respond(w, Response{
			Message: "Could not retrieve transactions",
			Error:   fmt.Errorf("parse query parameters: %w", err).Error(),
		}, http.StatusBadRequest,
			requestId)
		h.logs.Errorw("failed to parse query parameters", "error", err, "handler", GetTransactions, "request_id", requestId)
		return
	}

	txRequest := payload.TransactionsRequest{
		Transactions: values["transactionHashes"],
	}
	if err := txRequest.Validate(); err != nil {
		h.respond(w, Response{
			Message: "Request failed",
			Error:   fmt.Errorf("validate request payload: %w", err).Error(),
		}, http.StatusBadRequest,
			requestId)
		h.logs.Errorw("failed to validate request payload",
			"error", err,
			"handler", GetTransactions,
			"request_id", requestId)
		return
	}

	h.logs.Infow("transactions request received",
		"transactions", txRequest.Transactions,
		"handler", GetTransactions,
		"request_id", requestId)

	transactions, err := h.fethcher.GetTransactions(r.Context(), txRequest.Transactions)
	if err != nil {
		h.respond(w, Response{
			Message: "Could not retrieve transactions",
			Error:   fmt.Errorf("get transactions: %w", err).Error(),
		}, http.StatusInternalServerError,
			requestId)
		h.logs.Errorw("failed to get transactions",
			"error", err,
			"handler", GetTransactions,
			"request_id", requestId)
		return
	}

	h.logs.Infow("transactions retrieved",
		"transactions", transactions,
		"handler", GetTransactions,
		"request_id", requestId)

	transactionHashes := make([]string, 0, len(transactions))
	for _, tx := range transactions {
		transactionHashes = append(transactionHashes, tx.TransactionHash)
	}

	// save to user history
	authToken := r.Header.Get("AUTH_TOKEN")
	if authToken != "" && len(transactionHashes) > 0 {
		go func() {
			err = h.fethcher.SaveUserTransactionsHistory(r.Context(), authToken, transactionHashes)
			if err != nil {
				h.logs.Errorw("failed to save user history",
					"error", err,
					"handler", GetTransactions,
					"request_id", requestId)
			} else {
				h.logs.Infow("user history saved successfully",
					"num_of_transactions", len(transactionHashes),
					"handler", GetTransactions,
					"request_id", requestId)
			}
		}()
	}

	resp := map[string][]core.TransactionRecord{
		"transactions": transactions,
	}

	h.respond(w, resp, http.StatusOK, requestId)
}
func (h *FethHandler) HandleGetTransactionsRLP(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	reqIdCtx := r.Context().Value(middleware.RequestIDKey)
	if reqIdCtx != nil {
		requestId = reqIdCtx.(string)
	}

	path := r.URL.Path
	prefix := "/lime/eth/"

	rlphex := strings.TrimPrefix(path, prefix)

	if rlphex == "" {
		h.respond(w, Response{
			Message: "Request failed",
			Error:   "rlp hash parameter is required",
		}, http.StatusBadRequest,
			requestId)
		h.logs.Errorw("missing rlpHash parameter",
			"handler", GetTransactionsRLP,
			"request_id", requestId)
		return
	}

	transactionHashes, err := h.fethcher.ParseRLP(rlphex)
	if err != nil {
		h.respond(w, Response{
			Message: "Request failed",
			Error:   fmt.Errorf("parse RLP parameter: %w", err).Error(),
		}, http.StatusInternalServerError,
			requestId)
		h.logs.Errorw("failed to parse RLP parameter",
			"error", err,
			"handler", GetTransactionsRLP,
			"request_id", requestId)
		return
	}

	h.logs.Infow("rlp request parsed successfully",
		"transactions", transactionHashes,
		"handler", GetTransactionsRLP,
		"request_id", requestId)

	transactionRequest := payload.TransactionsRequest{
		Transactions: transactionHashes,
	}
	if err = transactionRequest.Validate(); err != nil {
		h.respond(w, Response{
			Message: "Request failed",
			Error:   fmt.Errorf("validate request RLP parameter: %w", err).Error(),
		}, http.StatusBadRequest,
			requestId)
		h.logs.Errorw("failed to validate request RLP parameter",
			"error", err,
			"handler", GetTransactions,
			"request_id", requestId)
		return
	}

	transactions, err := h.fethcher.GetTransactions(r.Context(), transactionRequest.Transactions)
	if err != nil {
		h.respond(w, Response{
			Message: "Request failed",
			Error:   fmt.Errorf("get transactions by hash: %w", err).Error(),
		}, http.StatusBadRequest,
			requestId)
		h.logs.Errorw("failed to get transactions by hash",
			"error", err,
			"handler", GetTransactions,
			"request_id", requestId)
		return
	}

	txsFound := make([]string, 0, len(transactions))
	for _, tx := range transactions {
		txsFound = append(txsFound, tx.TransactionHash)
	}

	// save to user history
	authToken := r.Header.Get("AUTH_TOKEN")
	if authToken != "" && len(txsFound) > 0 {
		go func() {
			err = h.fethcher.SaveUserTransactionsHistory(r.Context(), authToken, txsFound)
			if err != nil {
				h.logs.Errorw("failed to save user history",
					"error", err,
					"handler", GetTransactions,
					"request_id", requestId)
			} else {
				h.logs.Infow("user history saved successfully",
					"num_of_transactions", len(transactionHashes),
					"handler", GetTransactions,
					"request_id", requestId)
			}
		}()
	}

	resp := map[string][]core.TransactionRecord{
		"transactions": transactions,
	}

	h.respond(w, resp, http.StatusOK, requestId)
}

func (h *FethHandler) HandleGetMyTransactions(w http.ResponseWriter, r *http.Request) {
	requestId := ""

	reqIdCtx := r.Context().Value(middleware.RequestIDKey)
	if reqIdCtx != nil {
		requestId = reqIdCtx.(string)
	}

	authToken := r.Header.Get("AUTH_TOKEN")
	if authToken == "" {
		h.respond(w, Response{
			Message: "Authentication failed",
			Error:   "AUTH_TOKEN header is required",
		}, http.StatusUnauthorized,
			requestId)
		h.logs.Errorw("missing AUTH_TOKEN header", "handler", GetMyTransactions, "request_id", requestId)
		return
	}

	h.logs.Infow("user transactions request received", "authToken", authToken, "handler", GetMyTransactions, "request_id", requestId)

	transactions, err := h.fethcher.GetUserTransactionsHistory(r.Context(), authToken)
	if err != nil {
		h.respond(w, Response{
			Message: "Failed to get user transactions",
			Error:   fmt.Errorf("get user transactions: %w", err).Error(),
		}, http.StatusInternalServerError,
			requestId)
		h.logs.Errorw("failed to get user transactions", "error", err, "handler", GetMyTransactions)
		return
	}

	resp := map[string][]core.TransactionRecord{
		"transactions": transactions,
	}

	h.respond(w, resp, http.StatusOK, requestId)
}

func (h *FethHandler) HandleGetAllTransactions(w http.ResponseWriter, r *http.Request) {
	requestId := ""
	reqIdCtx := r.Context().Value(middleware.RequestIDKey)
	if reqIdCtx != nil {
		requestId = reqIdCtx.(string)
	}

	transactions, err := h.fethcher.GetAllDBTransactions(r.Context())
	if err != nil {
		h.respond(w, Response{
			Message: "Request failed",
			Error:   fmt.Errorf("get all transactions: %w", err).Error(),
		}, http.StatusInternalServerError,
			requestId)
		h.logs.Errorw("failed to get all transactions",
			"error", err,
			"handler", GetAllTransactions,
			"request_id", requestId)
		return
	}

	h.logs.Infow("transactions retrieved from DB",
		"request_id", requestId,
		"handler", GetAllTransactions,
		"transactions", transactions,
	)

	resp := map[string][]core.TransactionRecord{
		"transactions": transactions,
	}

	h.respond(w, resp, http.StatusOK, requestId)
}

func (h *FethHandler) respond(w http.ResponseWriter, resp any, code int, requestId string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, oopsErr, http.StatusInternalServerError)
		h.logs.Errorw("failed to encode response",
			"error", err,
			"request_id", requestId)
	}
}
