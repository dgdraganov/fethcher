package cmd

import (
	"fethcher/internal/config"
	"fethcher/internal/core"
	"fethcher/internal/db"
	"fethcher/internal/http/handler"
	"fethcher/internal/http/payload"
	"fethcher/internal/http/server"
	"fethcher/internal/repository"
	"fethcher/pkg/jwt"
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

	// connect to the default DB as fethcher database does not exist yet
	dbConn, err := db.NewFethDB("host=db user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		logger.Errorw("failed to connect to database", "error", err)
		return err
	}
	// now 'fethcher' database exsists and we can connect to it
	dbConn, err = db.NewFethDB(config.DBConnectionString)
	if err != nil {
		logger.Errorw("failed to connect to database", "error", err)
		return err
	}

	// jwt service
	jwtService := jwt.NewJWTService([]byte(config.JWTSecret))

	// repository
	repo := repository.NewFethRepo(dbConn)
	err = repo.MigrateAndSeed("fetcher")
	if err != nil {
		logger.Errorw("failed to migrate and seed database", "error", err)
		return err
	}

	// fethcher
	fethcher := core.NewFethcher(logger, repo, jwtService)

	// handler
	limeHlr := handler.NewFethHandler(
		logger,
		payload.DecodeValidator{},
		fethcher,
	)

	// register routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /lime/authenticate", limeHlr.HandleAuthenticate)

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
