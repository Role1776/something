package service

import (
	"log/slog"
	"todoai/internal/config"
	"todoai/internal/repository"
	"todoai/pkg/jwt"
	"todoai/pkg/mail"
)

type Service struct {
	Auth  AuthService
	Lists ListsService
}

func NewService(repo *repository.Repository, log *slog.Logger, cfg *config.Config, sender mail.Sender, jwt jwt.JWT) *Service {
	return &Service{
		Auth:  NewAuthService(repo.Auth, log, cfg, sender, jwt),
		Lists: NewListsService(&repo.Lists),
	}
}
