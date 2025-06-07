package cmd

import (
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

	// now 'fethcher' database exsists and we can connect to it
	dbConn, err := db.NewPostgresDB(config.DBConnectionString)
	if err != nil {
		logger.Errorw("failed to connect to database", "error", err)
		return err
	}

	// jwt service
	jwtService := jwt.NewJWTService([]byte(config.JWTSecret))

	// repository
	repo := repository.NewTransactionRepository(dbConn)
	err = repo.MigrateAndSeed()
	if err != nil {
		logger.Errorw("failed to migrate and seed database", "error", err)
		return err
	}

	// infura client token should be set in environment variable INFURA_TOKEN
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/2a99b1970a934959abf51e0b7df0fd62")
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
	limeHlr := handler.NewFethHandler(
		logger,
		payload.Decoder{},
		fethcher,
	)

	// middleware
	mux := http.NewServeMux()
	hdlr := middleware.NewLoggingMiddleware(logger).Logging(mux)
	hdlr = middleware.NewRequestIDMiddleware().RequestID(hdlr)

	// register routes
	mux.HandleFunc(handler.Authenticate, limeHlr.HandleAuthenticate)
	mux.HandleFunc(handler.GetTransactions, limeHlr.HandleGetTransactions)
	mux.HandleFunc(handler.GetTransactionsRLP, limeHlr.HandleGetTransactionsRLP)
	mux.HandleFunc(handler.GetMyTransactions, limeHlr.HandleGetMyTransactions)
	mux.HandleFunc(handler.GetAllTransactions, limeHlr.HandleGetAllTransactions)

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

	shErr := server.Shutdown()
	if err == nil {
		return shErr
	}

	return err
}
