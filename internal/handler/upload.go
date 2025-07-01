package handler

import (
	"errors"
	"io"
	"net/http"
	"todoai/internal/models"
	"todoai/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *handler) uploadFile(c *gin.Context) {
	mode := c.DefaultPostForm("mode", "default")
	if !isValidMode(mode) {
		newHTTPError(c, http.StatusBadRequest, "invalid mode")
		return
	}

	err := c.Request.ParseMultipartForm(100 << 20)
	if err != nil {
		newHTTPError(c, http.StatusBadRequest, "file upload failed")
		return
	}

	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		newHTTPError(c, http.StatusBadRequest, "file upload failed")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		newHTTPError(c, http.StatusInternalServerError, "file upload failed")
		return
	}

	text, err := h.service.Upload.ProcessFile(c.Request.Context(), fileBytes, handler.Filename, mode)
	if err != nil {
		if errors.Is(err, service.ErrFileTooLarge) {
			newHTTPError(c, http.StatusBadRequest, service.ErrFileTooLarge.Error())
			return
		}
		newHTTPError(c, http.StatusInternalServerError, "file upload failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"text": text})
}

func (h *handler) uploadText(c *gin.Context) {
	var document models.Document
	if err := c.ShouldBindJSON(&document); err != nil {
		newHTTPError(c, http.StatusBadRequest, "invalid response")
		return
	}

	if !isValidMode(document.Mode) {
		newHTTPError(c, http.StatusBadRequest, "invalid mode")
		return
	}

	text, err := h.service.Upload.ProcessText(c.Request.Context(), document.Text, document.Mode)
	if err != nil {
		if errors.Is(err, service.ErrFileTooLarge) {
			newHTTPError(c, http.StatusBadRequest, service.ErrFileTooLarge.Error())
			return
		}
		newHTTPError(c, http.StatusInternalServerError, "file upload text")
		return
	}

	c.JSON(http.StatusOK, gin.H{"text": text})
}

func isValidMode(mode string) bool {
	switch mode {
	case "default":
		return true
	case "book":
		return true
	case "doc":
		return true
	case "article":
		return true
	case "jurisprudence":
		return true
	default:
		return false
	}
}
