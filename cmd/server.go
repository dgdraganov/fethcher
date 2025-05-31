package cmd

import (
	"fethcher/internal/config"
	"fethcher/internal/http/handler"
	"fethcher/internal/http/server"
	"fethcher/pkg/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap/zapcore"
)

func Start() error {
	logger := log.NewZapLogger("fethcher", zapcore.InfoLevel)
	config, err := config.NewApp()
	if err != nil {
		logger.Errorw("failed to get config", "error", err)
		return err
	}

	mux := http.NewServeMux()

	limeHlr := handler.NewLimeHandler(logger)

	// register routes
	mux.HandleFunc("GET /lime/eth", limeHlr.HandleGetTransactions)

	srv := server.NewHTTP(logger, mux, config.Port)
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
