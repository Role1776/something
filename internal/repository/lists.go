package repository

type Lists interface {
}

type listsRepo struct{}

func newListsRepo() *listsRepo {
	return &listsRepo{}
}
