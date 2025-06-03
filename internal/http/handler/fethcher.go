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

type FethHandler struct {
	logs     *zap.SugaredLogger
	fethcher TransactionService
}

func NewFethHandler(logger *zap.SugaredLogger, transactionService TransactionService) *FethHandler {
	return &FethHandler{
		logs:     logger,
		fethcher: transactionService,
	}
}

func (h *FethHandler) HandleAuthenticate(w http.ResponseWriter, r *http.Request) {
	var request payload.AuthRequest
	if err := payload.DecodePayload(r, &request); err != nil {
		h.respond(w, Response{
			Message: "Could not authenticate you",
			Error:   fmt.Errorf("decoding payload: %w", err).Error(),
		}, http.StatusBadRequest)
		return
	}

	if err := request.Validate(); err != nil {
		h.respond(w, Response{
			Message: "Could not authenticate you",
			Error:   fmt.Errorf("validating payload: %w", err).Error(),
		}, http.StatusBadRequest)
		return
	}

	token, err := h.fethcher.Authenticate(request.ToCoreAuthMessage())
	if err != nil {
		// ErrUserNotFound
		resp := Response{
			Message: "Login failed",
		}

		var httpCode int

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

func (h *FethHandler) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	// not implemented
}

func (h *FethHandler) respond(w http.ResponseWriter, resp any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, oopsErr, http.StatusInternalServerError)
		h.logs.Errorw("failed to encode response", "error", err, "handler", Authenticate)
	}
}
