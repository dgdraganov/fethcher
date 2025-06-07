package server

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type HTTPServer struct {
	server *http.Server
	logs   *zap.SugaredLogger
}

func NewHTTP(logger *zap.SugaredLogger, mux http.Handler, port string) *HTTPServer {
	server := &http.Server{
		Handler: mux,
		Addr:    fmt.Sprintf(":%s", port),
	}
	return &HTTPServer{
		server: server,
		logs:   logger,
	}
}

func (s *HTTPServer) Run() <-chan error {
	s.logs.Infow(
		"service starting",
		"app_port", s.server.Addr,
	)

	errChan := make(chan error)
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			s.logs.Errorw("server shut down", "error", err)
			errChan <- err
		}
	}()
	return errChan
}

func (s *HTTPServer) Shutdown() error {
	s.logs.Info("shutting down server...")

	if err := s.server.Shutdown(context.Background()); err != nil {
		s.logs.Error(
			"server shutdown failed",
			"error", err,
		)
		return err
	}
	return nil
}
