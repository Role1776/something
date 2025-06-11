package handler

import (
	"todoai/internal/service"
	"todoai/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service *service.Service
	jwt     jwt.JWT
}

func NewHandler(service *service.Service, jwt jwt.JWT) *handler {
	return &handler{service: service, jwt: jwt}
}

func (h *handler) HandlerRegistrator() *gin.Engine {
	r := gin.Default()
	r.Use(h.errorHandler())
	auth := r.Group("/auth")
	{
		auth.POST("/sign-in", h.signIn)
		auth.POST("/sign-up", h.signUp)
		auth.POST("/verify/:code", h.verify)
		auth.POST("/refresh", h.refreshToken)
		auth.POST("/logout", h.logout)
	}

	api := r.Group("/api", h.authMiddleware)
	{
		lists := api.Group("/lists")
		{
			lists.GET("/get", h.getLists)
			lists.GET("/get/:id", h.getListById)
			lists.POST("/create", h.createList)
			lists.PUT("/update", h.updateList)
			lists.DELETE("/delete", h.deleteList)
		}

	}

	return r
}
