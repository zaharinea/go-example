package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/zaharinea/go-example/config"
	_ "github.com/zaharinea/go-example/docs"
	"github.com/zaharinea/go-example/pkg/service"
)

// Handler struct
type Handler struct {
	config   *config.Config
	services *service.Service
}

// NewHandler returns a new Handler struct
func NewHandler(config *config.Config, services *service.Service) *Handler {
	return &Handler{config: config, services: services}
}

// InitRouter return a initialized Router
func (h *Handler) InitRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	url := ginSwagger.URL("/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	r.GET("/api/healthcheck", h.Healthcheck)
	r.POST("/api/users", h.CreateUser)
	r.GET("/api/users", h.ListUsers)
	r.GET("/api/users/:id", h.GetUserByID)
	r.PUT("/api/users/:id", h.UpdateUser)
	r.DELETE("/api/users/:id", h.DeleteUserByID)
	return r
}
