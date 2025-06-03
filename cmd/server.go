package cmd

import (
	"fethcher/internal/config"
	"fethcher/internal/core"
	"fethcher/internal/db"
	"fethcher/internal/http/handler"
	"fethcher/internal/http/server"
	"fethcher/internal/storage"
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
	config, err := config.NewConfig()
	if err != nil {
		logger.Errorw("failed to create config", "error", err)
		return err
	}

	// // connect to the default DB as fethcher database does not exist yet
	// dbConn, err := db.NewGormDB("host=db user=postgres password=postgres dbname=postgres sslmode=disable")
	// if err != nil {
	// 	logger.Errorw("failed to connect to database", "error", err)
	// 	return err
	// }

	dbConn, err := db.NewGormDB(config.DBConnectionString)
	if err != nil {
		logger.Errorw("failed to connect to database", "error", err)
		return err
	}

	jwtService := jwt.NewService([]byte(config.JWTSecret))

	repo := storage.NewUserRepository(dbConn)

	if err := repo.MigrateAndSeed("fetcher"); err != nil {
		logger.Errorw("failed to migrate and seed database", "error", err)
		return err
	}

	fethcher := core.NewFethcher(logger, repo, jwtService)

	limeHlr := handler.NewFethHandler(
		logger,
		fethcher,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /lime/authenticate", limeHlr.HandleAuthenticate)

	srv := server.NewHTTP(logger, mux, config.Port)
	return run(srv)
}

func run(server *server.HTTPServer) error {
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
