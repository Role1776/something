package service

import (
	"database/sql"
	"log/slog"
	"todoai/internal/config"
	"todoai/internal/gateway/ai"
	"todoai/pkg/jwt"
	"todoai/pkg/mail"
)

type Service struct {
	Auth  AuthService
	Lists ListsService
	File  File
}

func NewService(db *sql.DB, tm *TransactionManager, log *slog.Logger, cfg *config.Config, sender mail.Sender, jwt jwt.JWT, ai ai.AI) *Service {
	return &Service{
		Auth:  NewAuthService(tm, log, cfg, sender, jwt),
		Lists: NewListsService(tm, log),
		File:  NewFileService(cfg, log, ai),
	}
}
