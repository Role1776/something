package service

import (
	"context"
	"database/sql"
	"fmt"
	"todoai/internal/repository"
)

type TransactionManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) NewAuthRepo() repository.Auth {
	return repository.NewAuthRepo(tm.db)
}

func (tm *TransactionManager) NewListsRepo() repository.Lists {
	return repository.NewListsRepo(tm.db)
}

type TransactionalRepos struct {
	Auth  repository.Auth
	Lists repository.Lists
}

func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(repos *TransactionalRepos) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	transactionalRepo := &TransactionalRepos{
		Auth:  repository.NewAuthRepo(tx),
		Lists: repository.NewListsRepo(tx),
	}

	if err := fn(transactionalRepo); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
