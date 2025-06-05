package handler

import (
	"encoding/json"
	"errors"
	"fethcher/internal/core"
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

type fethHandler struct {
	logs             *zap.SugaredLogger
	requestValidator RequestValidator
	fethcher         TransactionService
}

func NewFethHandler(logger *zap.SugaredLogger, requestValidator RequestValidator, transactionService TransactionService) *fethHandler {
	return &fethHandler{
		logs:             logger,
		requestValidator: requestValidator,
		fethcher:         transactionService,
	}
}

func (h *fethHandler) HandleAuthenticate(w http.ResponseWriter, r *http.Request) {
	var payload payload.AuthRequest
	if err := h.requestValidator.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		h.respond(w, Response{
			Message: "Could not authenticate you",
			Error:   fmt.Errorf("invalid request payload: %w", err).Error(),
		}, http.StatusBadRequest)
		h.logs.Errorw("failed to decode and validate request payload", "error", err, "handler", Authenticate)
		return
	}

	token, err := h.fethcher.Authenticate(r.Context(), payload.ToMessage())
	if err != nil {
		resp := Response{
			Message: "Login failed",
		}
		httpCode := http.StatusInternalServerError
		if errors.Is(err, core.ErrUserNotFound) {
			resp.Error = err.Error()
			httpCode = http.StatusUnauthorized
		} else if errors.Is(err, core.ErrIncorrectPassword) {
			httpCode = http.StatusUnauthorized
			resp.Error = err.Error()
		} else {
			httpCode = http.StatusInternalServerError
			resp.Error = "unexpected error occurred"
		}

		h.respond(w, resp, httpCode)
		h.logs.Errorw("authentication failed", "error", err, "handler", Authenticate)
		return
	}

	resp := map[string]string{
		"token": token,
	}
	h.respond(w, resp, http.StatusOK)
}

func (h *fethHandler) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		h.respond(w, Response{
			Message: "Failed to parse query parameters",
			Error:   fmt.Errorf("parse query parameters: %w", err).Error(),
		}, http.StatusBadRequest)
		h.logs.Errorw("failed to parse query parameters", "error", err, "handler", GetTransactions)
		return
	}

	// Get all values for the "transactionHashes" key
	transactionHashes := values["transactionHashes"]

	h.logs.Infow("transactions request received", "num_of_transactions", len(transactionHashes), "handler", GetTransactions)

	transactions, err := h.fethcher.GetTransactions(r.Context(), transactionHashes)
	if err != nil {
		h.respond(w, Response{
			Message: "Failed to get transactions",
			Error:   fmt.Errorf("get transactions: %w", err).Error(),
		}, http.StatusInternalServerError)
		h.logs.Errorw("failed to get transactions", "error", err, "handler", GetTransactions)
		return
	}

	transactionHashes = make([]string, 0, len(transactions))
	for _, tx := range transactions {
		transactionHashes = append(transactionHashes, tx.TransactionHash)
	}

	// save to user history
	authToken := r.Header.Get("AUTH_TOKEN")
	if authToken != "" {
		err = h.fethcher.SaveUserTransactionsHistory(r.Context(), authToken, transactionHashes)
		if err != nil {
			h.logs.Errorw("failed to save user history", "error", err, "handler", GetTransactions)
		} else {
			h.logs.Infow("user history saved successfully", "num_of_transactions", len(transactionHashes), "handler", GetTransactions)
		}
	}

	resp := map[string][]core.TransactionRecord{
		"transactions": transactions,
	}

	h.respond(w, resp, http.StatusOK)
}
func (h *fethHandler) HandleGetTransactionsRLP(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	prefix := "/lime/eth/"

	rlphex := strings.TrimPrefix(path, prefix)

	if rlphex == "" {
		h.respond(w, Response{
			Message: "Request failed",
			Error:   "rlp hash parameter is required",
		}, http.StatusBadRequest)
		h.logs.Errorw("missing rlpHash parameter", "handler", GetTransactionsRLP)
		return
	}

	h.logs.Infow("transactions RLP request received", "rlpHash", rlphex, "handler", GetTransactionsRLP)

	transactions, err := h.fethcher.GetTransactionsRLP(r.Context(), rlphex)
	if err != nil {
		h.respond(w, Response{
			Message: "Failed to get transactions RLP",
			Error:   fmt.Errorf("get transactions RLP: %w", err).Error(),
		}, http.StatusInternalServerError)
		h.logs.Errorw("failed to get transactions RLP", "error", err, "handler", GetTransactionsRLP)
		return
	}

	if len(transactions) == 0 {
		h.respond(w, Response{
			Message: "Request failed",
			Error:   "no transactions found",
		}, http.StatusNotFound)
		h.logs.Infow("no transactions found for the provided RLP hash", "rlpHash", rlphex, "handler", GetTransactionsRLP)
		return
	}

	resp := map[string][]core.TransactionRecord{
		"transactions": transactions,
	}

	h.respond(w, resp, http.StatusOK)
}

func (h *fethHandler) respond(w http.ResponseWriter, resp any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, oopsErr, http.StatusInternalServerError)
		h.logs.Errorw("failed to encode response", "error", err, "handler", Authenticate)
	}
}
