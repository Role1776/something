package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"todoai/internal/config"
	"todoai/internal/handler"
	"todoai/internal/repository"
	"todoai/internal/server"
	"todoai/internal/service"
	"todoai/pkg/jwt"
	smpt "todoai/pkg/mail/smtp"
)

type App struct {
	Server *server.HttpServer
	db     *sql.DB
}

func NewApp(cfg *config.Config, log *slog.Logger) (*App, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Database.ConnTimeout)
	defer cancel()

	postgresCfg := repository.PostgresConfig{
		ConnString: cfg.Database.ConnString,
	}

	db, err := repository.ConnectToPostgres(ctx, &postgresCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	sender, err := smpt.NewSMTPSender(cfg.Email.SMTPFromAddress, cfg.Email.SMTPPassword, cfg.Email.SMTPHost, cfg.Email.SMTPPort)
	if err != nil {
		return nil, fmt.Errorf("create smtp sender: %w", err)
	}
	jwt := jwt.NewJWT()

	repo := repository.NewRepository(db)
	service := service.NewService(repo, log, cfg, sender, jwt)
	handler := handler.NewHandler(service, jwt)
	router := handler.HandlerRegistrator()

	server := server.NewServer(cfg, router)

	return &App{
		Server: server,
		db:     db,
	}, nil
}

func (a *App) Close() error {
	return a.db.Close()
}
