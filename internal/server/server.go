package server

import (
	"context"
	"net/http"
	"todoai/internal/config"
)

type HttpServer struct {
	HttpServer *http.Server
}

func NewServer(cfg *config.Config, handler http.Handler) *HttpServer {
	return &HttpServer{
		HttpServer: &http.Server{
			Addr:           ":" + cfg.Server.Port,
			Handler:        handler,
			MaxHeaderBytes: cfg.Server.MaxHeaderMegabytes << 20,
			WriteTimeout:   cfg.Server.WriteTimeout,
			ReadTimeout:    cfg.Server.ReadTimeout,
		},
	}
}

func (s *HttpServer) Run() error {
	return s.HttpServer.ListenAndServe()
}

func (s *HttpServer) Stop(ctx context.Context) error {
	return s.HttpServer.Shutdown(ctx)
}
