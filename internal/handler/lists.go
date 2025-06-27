package handler

import (
	"net/http"
	"strconv"
	"todoai/internal/models"

	"github.com/gin-gonic/gin"
)

func (h *handler) getListById(c *gin.Context) {
	listID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newHTTPError(c, http.StatusBadRequest, "invalid list id")
		return
	}
	userID := c.GetInt("userID")
	list, err := h.service.Lists.GetByID(c.Request.Context(), listID, userID)
	if err != nil {
		newHTTPError(c, http.StatusInternalServerError, "failed to get list by id")
		return
	}
	c.JSON(http.StatusOK, list)
}
func (h *handler) getLists(c *gin.Context) {
	userID := c.GetInt("userID")
	lists, err := h.service.Lists.Get(c.Request.Context(), userID)
	if err != nil {
		newHTTPError(c, http.StatusInternalServerError, "failed to get lists")
		return
	}
	c.JSON(http.StatusOK, lists)
}

func (h *handler) createList(c *gin.Context) {
	var list models.List
	if err := c.ShouldBindJSON(&list); err != nil {
		newHTTPError(c, http.StatusBadRequest, "invalid list data")
		return
	}
	userID := c.GetInt("userID")
	err := h.service.Lists.Create(c.Request.Context(), &list, userID)
	if err != nil {
		newHTTPError(c, http.StatusInternalServerError, "failed to create list")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "list created successfully"})
}

func (h *handler) updateList(c *gin.Context) {
	var list models.List
	if err := c.ShouldBindJSON(&list); err != nil {
		newHTTPError(c, http.StatusBadRequest, "invalid list data")
		return
	}

	userID := c.GetInt("userID")
	listID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newHTTPError(c, http.StatusBadRequest, "invalid list id")
		return
	}

	err = h.service.Lists.UpdateOnlyBody(c.Request.Context(), listID, list.Body, userID)
	if err != nil {
		newHTTPError(c, http.StatusInternalServerError, "failed to update list")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "list updated successfully"})
}

func (h *handler) deleteList(c *gin.Context) {
	listID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newHTTPError(c, http.StatusBadRequest, "invalid list id")
		return
	}
	userID := c.GetInt("userID")
	err = h.service.Lists.Delete(c.Request.Context(), listID, userID)
	if err != nil {
		newHTTPError(c, http.StatusInternalServerError, "failed to delete list")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "list deleted successfully"})
}
