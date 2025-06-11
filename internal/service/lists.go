package service

import "todoai/internal/repository"

type ListsService interface{}

type listsService struct {
	lists repository.Lists
}

func NewListsService(lists repository.Lists) *listsService {
	return &listsService{lists: lists}
}
