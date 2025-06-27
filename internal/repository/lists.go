package repository

import (
	"context"
	"fmt"
	"todoai/internal/models"
)

type Lists interface {
	Create(ctx context.Context, list *models.List, userID int) error
	UpdateBody(ctx context.Context, listID int, body string, userID int) error
	Delete(ctx context.Context, listID int, userID int) error
	GetByID(ctx context.Context, listID int, userID int) (models.List, error)
	Get(ctx context.Context, userID int) ([]models.List, error)
}

type listsRepo struct {
	db Querier
}

func NewListsRepo(db Querier) *listsRepo {
	return &listsRepo{db: db}
}

func (l *listsRepo) Create(ctx context.Context, list *models.List, userID int) error {
	const op = "repository.CreateList"

	const query = `
		INSERT INTO lists (title, body, user_id)
		VALUES ($1, $2, $3)
	`
	_, err := l.db.ExecContext(ctx, query, list.Title, list.Body, userID)
	if err != nil {

		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (l *listsRepo) GetByID(ctx context.Context, listID int, userID int) (models.List, error) {
	const op = "repository.GetListByID"

	const query = `
		SELECT id, title, body FROM lists
		WHERE id = $1 AND user_id = $2
	`
	rows := l.db.QueryRowContext(ctx, query, listID, userID)

	var list models.List
	if err := rows.Scan(&list.ID, &list.Title, &list.Body); err != nil {
		return models.List{}, fmt.Errorf("%s: %w", op, err)
	}
	return list, nil
}

func (l *listsRepo) Get(ctx context.Context, userID int) ([]models.List, error) {
	const op = "repository.GetLists"

	const query = `
		SELECT id, title, body FROM lists
		WHERE user_id = $1
	`
	rows, err := l.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var lists []models.List
	for rows.Next() {
		var list models.List
		if err := rows.Scan(&list.ID, &list.Title, &list.Body); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		lists = append(lists, list)
	}
	return lists, nil
}

func (l *listsRepo) UpdateBody(ctx context.Context, listID int, body string, userID int) error {
	const op = "repository.UpdateListBody"

	const query = `
		UPDATE lists
		SET body = $1
		WHERE id = $2 AND user_id = $3
	`
	_, err := l.db.ExecContext(ctx, query, body, listID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (l *listsRepo) Delete(ctx context.Context, listID int, userID int) error {
	const op = "repository.DeleteList"

	const query = `
		DELETE FROM lists
		WHERE id = $1 AND user_id = $2
	`
	_, err := l.db.ExecContext(ctx, query, listID, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
