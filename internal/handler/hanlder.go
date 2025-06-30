package handler

import (
	"todoai/internal/service"
	"todoai/pkg/jwt"

	"github.com/gin-contrib/cors"
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

	r.Use(gin.Recovery())
	r.Use(cors.Default())
	r.Use(h.errorHandler())

	auth := r.Group("/auth")
	{
		auth.POST("/sign-in", h.signIn)
		auth.POST("/sign-up", h.signUp)
		auth.POST("/verify", h.verify)
		auth.POST("/resend-code", h.resend_code)
		auth.POST("/refresh", h.refreshToken)
		auth.POST("/logout", h.logout)
	}

	api := r.Group("/api")
	{
		lists := api.Group("/lists", h.authMiddleware)
		{
			lists.GET("/get", h.getLists)
			lists.GET("/get/:id", h.getListById)
			lists.POST("/create", h.createList)
			lists.PATCH("/update/:id", h.updateList)
			lists.DELETE("/delete/:id", h.deleteList)
		}

	}
	upload := api.Group("/upload")
	{
		upload.POST("/file", h.uploadFile)
		upload.POST("/text", h.uploadText)
	}

	return r
}
