package handler

import (
	"errors"
	"net/http"
	"todoai/internal/models"
	"todoai/internal/repository"
	"todoai/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *handler) signIn(c *gin.Context) {
	var authData models.SecondAuth
	if err := c.ShouldBindJSON(&authData); err != nil {
		newHTTPError(c, http.StatusBadRequest, "invalid response")
		return
	}

	tokens, err := h.service.Auth.SignIn(c.Request.Context(), &authData)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			newHTTPError(c, http.StatusUnauthorized, "user not found")
			return
		}
		newHTTPError(c, http.StatusInternalServerError, "failed to authenticate user")
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *handler) signUp(c *gin.Context) {
	var authData models.FirstAuth
	if err := c.ShouldBindJSON(&authData); err != nil {
		newHTTPError(c, http.StatusBadRequest, "invalid response")
		return
	}

	err := h.service.Auth.SignUp(c.Request.Context(), &authData)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			c.JSON(http.StatusOK, gin.H{"message": "confirmation email has been sent"})
			return
		}
		newHTTPError(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})

}

func (h *handler) verify(c *gin.Context) {
	var code models.VerificationCode
	if err := c.ShouldBindJSON(&code); err != nil {
		newHTTPError(c, http.StatusBadRequest, "invalid response")
		return
	}

	if err := h.service.Auth.VerifyUser(c.Request.Context(), code.Code); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			newHTTPError(c, http.StatusBadRequest, "invalid response")
			return
		}
		newHTTPError(c, http.StatusInternalServerError, "failed to verify user")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user verified successfully"})
}

func (h *handler) refreshToken(c *gin.Context) {
	var refreshToken models.RefreshToken
	if err := c.ShouldBindJSON(&refreshToken); err != nil {
		newHTTPError(c, http.StatusBadRequest, "invalid response")
		return
	}

	tokens, err := h.service.Auth.RefreshToken(c.Request.Context(), refreshToken.RefreshToken)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			newHTTPError(c, http.StatusUnauthorized, "token is dead")
			return
		}
		if errors.Is(err, service.ErrTokenExpired) {
			newHTTPError(c, http.StatusUnauthorized, "token expired")
			return
		}
		newHTTPError(c, http.StatusInternalServerError, "failed to refresh token")
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *handler) logout(c *gin.Context) {
	var refreshToken models.RefreshToken
	if err := c.ShouldBindJSON(&refreshToken); err != nil {
		newHTTPError(c, http.StatusBadRequest, "invalid response")
		return
	}

	if err := h.service.Auth.Logout(c.Request.Context(), refreshToken.RefreshToken); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			newHTTPError(c, http.StatusUnauthorized, "token is dead")
			return
		}
		newHTTPError(c, http.StatusInternalServerError, "failed to logout")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user logged out successfully"})
}
