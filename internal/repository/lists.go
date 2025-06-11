package repository

type Lists interface {
}

type listsRepo struct {
	db Querier
}

func NewListsRepo(db Querier) *listsRepo {
	return &listsRepo{db: db}
}
