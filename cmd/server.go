package cmd

import (
	"context"
	"fethcher/internal/config"
	"fethcher/internal/core"
	"fethcher/internal/db"
	"fethcher/internal/ethereum"
	"fethcher/internal/http/handler"
	"fethcher/internal/http/handler/middleware"
	"fethcher/internal/http/payload"
	"fethcher/internal/http/server"
	"fethcher/internal/repository"
	"fethcher/pkg/jwt"
	"fethcher/pkg/log"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap/zapcore"
)

func Start() error {
	logger := log.NewZapLogger("fethcher", zapcore.InfoLevel)

	config, err := config.NewAppConfig()
	if err != nil {
		logger.Errorw("failed to create config", "error", err)
		return err
	}

	dbConn, err := db.NewPostgresDB(config.DBConnectionString)
	if err != nil {
		logger.Errorw("failed to connect to database", "error", err)
		return err
	}

	// jwt service
	jwtService := jwt.NewJWTService([]byte(config.JWTSecret))

	// repository
	repo := repository.NewTransactionRepository(dbConn)

	err = repo.MigrateTables(
		&repository.Transaction{},
		&repository.User{},
		&repository.UserTransaction{})
	if err != nil {
		logger.Errorw("failed to migrate tables to database", "error", err)
		return err
	}

	err = repo.SeedUserTable(context.Background())
	if err != nil {
		logger.Errorw("failed to seed user table", "error", err)
		return err
	}

	client, err := ethclient.Dial(config.NodeURL)
	if err != nil {
		logger.Errorw("infura connection failed", "error", err)
		return err
	}

	ethService := ethereum.NewEthService(client)

	// fethcher
	fethcher := core.NewFethcher(
		logger,
		repo,
		jwtService,
		ethService)

	// handler
	fethHlr := handler.NewFethHandler(
		logger,
		payload.Decoder{},
		fethcher)

	// middleware
	mux := http.NewServeMux()
	hdlr := middleware.NewLoggingMiddleware(logger).Logging(mux)
	hdlr = middleware.NewRequestIDMiddleware().RequestID(hdlr)

	// register routes
	mux.HandleFunc(handler.Authenticate, fethHlr.HandleAuthenticate)
	mux.HandleFunc(handler.GetTransactions, fethHlr.HandleGetTransactions)
	mux.HandleFunc(handler.GetTransactionsRLP, fethHlr.HandleGetTransactionsRLP)
	mux.HandleFunc(handler.GetMyTransactions, fethHlr.HandleGetMyTransactions)
	mux.HandleFunc(handler.GetAllTransactions, fethHlr.HandleGetAllTransactions)

	srv := server.NewHTTP(logger, hdlr, config.Port)
	return run(srv)
}

func run(server *server.HTTPServer) error {
	// expect a signal to gracefully shutdown the server
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	errChan := server.Run()

	var err error
	select {
	case <-sig:
	case err = <-errChan:
	}

	sdErr := server.Shutdown()
	if err == http.ErrServerClosed && sdErr != nil {
		return fmt.Errorf("server shutdown: %w", sdErr)
	}

	return err
}
