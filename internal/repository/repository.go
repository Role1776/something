package repository

import (
	"database/sql"
)

type Repository struct {
	Auth  Auth
	Lists Lists
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Auth:  newAuthRepo(db),
		Lists: newListsRepo(),
	}
}
