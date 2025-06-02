package handler

import (
	"encoding/json"
	"errors"
	"fethcher/internal/core"
	"fethcher/internal/http/payload"
	"fmt"
	"net/http"

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
		return
	}

	token, err := h.fethcher.Authenticate(payload.ToMessage())
	if err != nil {
		// ErrUserNotFound
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
		return
	}

	resp := map[string]string{
		"token": token,
	}
	h.respond(w, resp, http.StatusOK)
}

func (h *fethHandler) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	// not implemented
}

func (h *fethHandler) respond(w http.ResponseWriter, resp any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, oopsErr, http.StatusInternalServerError)
		h.logs.Errorw("failed to encode response", "error", err, "handler", Authenticate)
	}
}
