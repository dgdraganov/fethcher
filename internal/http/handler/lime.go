package handler

import (
	"net/http"

	"go.uber.org/zap"
)

// mux.HandleFunc("/lime/eth", handler.HandleGetTransactions)
// mux.HandleFunc("/lime/eth/", handler.HandleGetTransactionsFromRLP)
// mux.HandleFunc("/lime/all", handler.HandleGetAllTransactions)
// mux.HandleFunc("/lime/authenticate", handler.HandleAuthenticate)
// mux.HandleFunc("/lime/my", handler.HandleGetMyTransactions)

var (
	GetTransactions    = "GET /lime/eth"
	GetTransactionsRLP = "GET /lime/eth/{rlpHash}"
	GetAllTransactions = "GET /lime/all"
	Authenticate       = "POST /lime/authenticate"
	GetMyTransactions  = "GET /lime/my"
)

type limeHandler struct {
	logs *zap.SugaredLogger
}

func NewLimeHandler(logger *zap.SugaredLogger) *limeHandler {
	return &limeHandler{
		logs: logger,
	}
}
func (h *limeHandler) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	// not implemented
}
