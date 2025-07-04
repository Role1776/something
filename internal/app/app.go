package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"todoai/internal/config"
	"todoai/internal/gateway/ai"
	"todoai/internal/handler"
	"todoai/internal/repository"
	"todoai/internal/server"
	"todoai/internal/service"
	"todoai/pkg/jwt"
	smpt "todoai/pkg/mail/smtp"

	"github.com/google/generative-ai-go/genai"
	"golang.org/x/time/rate"
)

type App struct {
	server   *server.HttpServer
	aIclient *genai.Client
	db       *sql.DB
	log      *slog.Logger
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

	aiClient, err := ai.NewClient(ctx, cfg.AI.Key)
	if err != nil {
		return nil, fmt.Errorf("create ai client: %w", err)
	}

	ratelimiterAI := rate.NewLimiter(rate.Limit(1), 1)

	ai := ai.NewAI(cfg.AI.Model, aiClient, ratelimiterAI, log)
	tm := service.NewTransactionManager(db)
	service := service.NewService(db, tm, log, cfg, sender, jwt, ai)
	handler := handler.NewHandler(service, jwt)
	router := handler.HandlerRegistrator()

	server := server.NewServer(cfg, router)

	return &App{
		server:   server,
		db:       db,
		aIclient: aiClient,
		log:      log,
	}, nil
}

func (a *App) Run() error {
	serverErr := make(chan error, 1)

	go func() {
		a.log.Info("server is starting", "addr", a.server.HttpServer.Addr)
		if err := a.server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		a.log.Info("Shutting down server...")
	case serverErr := <-serverErr:
		a.log.Error("Server run failed", slog.Any("error", serverErr))
		return serverErr
	}
	return nil
}

func (a *App) Close(ctx context.Context) error {
	var errs []error

	if err := a.server.Stop(ctx); err != nil {
		errs = append(errs, fmt.Errorf("http server close: %w", err))
	}

	if err := a.aIclient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("ai client close: %w", err))
	}

	if err := a.db.Close(); err != nil {
		errs = append(errs, fmt.Errorf("db close: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
