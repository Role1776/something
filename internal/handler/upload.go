package handler

import (
	"errors"
	"io"
	"net/http"
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

	text, err := h.service.File.ProcessFile(c.Request.Context(), fileBytes, handler.Filename, mode)
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
	c.JSON(http.StatusOK, gin.H{"text": "text uploaded successfully"})
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
