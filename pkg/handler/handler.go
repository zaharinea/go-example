package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/zaharinea/go-example/config"
	_ "github.com/zaharinea/go-example/docs"
	"github.com/zaharinea/go-example/pkg/repository"
	"github.com/zaharinea/go-example/pkg/service"
)

// Handler struct
type Handler struct {
	config   *config.Config
	repos    *repository.Repository
	services *service.Service
}

// NewHandler returns a new Handler struct
func NewHandler(config *config.Config, repos *repository.Repository, services *service.Service) *Handler {
	return &Handler{config: config, repos: repos, services: services}
}

// InitRoutes initialize endpoint
func (h *Handler) InitRoutes(app *gin.Engine) {
	url := ginSwagger.URL("/swagger/doc.json")
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	app.GET("/api/healthcheck", h.Healthcheck)
	app.POST("/api/users", h.CreateUser)
	app.GET("/api/users", h.ListUsers)
	app.GET("/api/users/:id", h.GetUserByID)
	app.PUT("/api/users/:id", h.UpdateUser)
	app.DELETE("/api/users/:id", h.DeleteUserByID)
}
