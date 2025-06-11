package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *handler) authMiddleware(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	parts := strings.Split(accessToken, " ")

	if len(parts) != 2 || parts[0] != "Bearer" {
		err := fmt.Errorf("expected Bearer token, got: %s", accessToken)
		logrus.Warnf("auth error: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid authorization header format",
		})
		return
	}

	id, err := h.jwt.ParseAccessToken(parts[1])
	if err != nil {
		logrus.Warnf("invalid token: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid token",
		})
		return
	}

	c.Set("userID", id)
	c.Next()
}
