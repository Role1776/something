package models

type Document struct {
	Mode string `json:"mode" binding:"required"`
	Text string `json:"text" binding:"required"`
}
