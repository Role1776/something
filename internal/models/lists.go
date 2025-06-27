package models

type List struct {
	ID    int    `json:"id"`
	Title string `json:"title" binding:"required,min=2,max=128"`
	Body  string `json:"body" binding:"required"`
}
