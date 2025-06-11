package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func newHTTPError(c *gin.Context, code int, message string) {
	c.Error(&HTTPError{
		Code:    code,
		Message: message,
	})
	c.AbortWithStatus(code)
}

func (h *handler) errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			switch e := err.Err.(type) {
			case *HTTPError:
				c.JSON(e.Code, gin.H{"error": e.Message})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		}
	}
}
