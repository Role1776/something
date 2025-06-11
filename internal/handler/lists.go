package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *handler) getListById(c *gin.Context) {
	c.Error(&HTTPError{
		Code:    404,
		Message: "Not implemented",
	})
}
func (h *handler) getLists(c *gin.Context) {
	// TODO: Implement logic to get all lists
	c.JSON(200, "HELLO")
}

func (h *handler) createList(c *gin.Context) {
	// TODO: Implement logic to create a new list
}

func (h *handler) updateList(c *gin.Context) {
	// TODO: Implement logic to update a list by ID
}

func (h *handler) deleteList(c *gin.Context) {
	// TODO: Implement logic to delete a list by ID
}
