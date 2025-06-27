package service

import (
	"context"
	"log/slog"
	"todoai/internal/models"
)

type ListsService interface {
	Create(ctx context.Context, list *models.List, userID int) error
	Get(ctx context.Context, userID int) ([]models.List, error)
	GetByID(ctx context.Context, listID int, userID int) (models.List, error)
	UpdateOnlyBody(ctx context.Context, listID int, text string, userID int) error
	Delete(ctx context.Context, listID int, userID int) error
}

type listsService struct {
	log *slog.Logger
	tm  *TransactionManager
}

func NewListsService(tm *TransactionManager, log *slog.Logger) *listsService {
	return &listsService{tm: tm, log: log}
}

func (s *listsService) Create(ctx context.Context, list *models.List, userID int) error {
	const op = "service.listsService.Create"
	repo := s.tm.NewListsRepo()
	err := repo.Create(ctx, list, userID)
	if err != nil {
		s.log.Error(op, "error", err)
		return err
	}
	return nil
}

func (s *listsService) Get(ctx context.Context, userID int) ([]models.List, error) {
	const op = "service.listsService.Get"
	repo := s.tm.NewListsRepo()
	lists, err := repo.Get(ctx, userID)
	if err != nil {
		s.log.Error(op, "error", err)
		return nil, err
	}
	return lists, nil
}

func (s *listsService) GetByID(ctx context.Context, listID int, userID int) (models.List, error) {
	const op = "service.listsService.GetByID"
	repo := s.tm.NewListsRepo()
	list, err := repo.GetByID(ctx, listID, userID)
	if err != nil {
		s.log.Error(op, "error", err)
		return models.List{}, err
	}
	return list, nil
}

func (s *listsService) UpdateOnlyBody(ctx context.Context, listID int, text string, userID int) error {
	const op = "service.listsService.UpdateOnlyBody"
	repo := s.tm.NewListsRepo()
	err := repo.UpdateBody(ctx, listID, text, userID)
	if err != nil {
		s.log.Error(op, "error", err)
		return err
	}
	return nil
}

func (s *listsService) Delete(ctx context.Context, listID int, userID int) error {
	const op = "service.listsService.Delete"
	repo := s.tm.NewListsRepo()
	err := repo.Delete(ctx, listID, userID)
	if err != nil {
		s.log.Error(op, "error", err)
		return err
	}
	return nil
}
