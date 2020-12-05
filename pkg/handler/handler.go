package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/zaharinea/go-example/config"
	_ "github.com/zaharinea/go-example/docs"
	"github.com/zaharinea/go-example/pkg/service"
)

const (
	errorMessageNotFound         = "Not found"
	errorMessageMethodNotAllowed = "Method not allowed"
	errorMessageServerError      = "Server error"
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

// InitRoutes initialize endpoint
// @title Example API
// @version 1.0
// @description This is an example http api server
func (h *Handler) InitRoutes(engine *gin.Engine) {
	engine.NoRoute(NoRouteHandler)
	engine.NoMethod(NoMethodHandler)

	url := ginSwagger.URL("/swagger/doc.json")
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	engine.GET("/api/healthcheck", h.Healthcheck)
	engine.POST("/api/users", h.CreateUser)
	engine.GET("/api/users", h.ListUsers)
	engine.GET("/api/users/:id", h.GetUserByID)
	engine.PUT("/api/users/:id", h.UpdateUser)
	engine.DELETE("/api/users/:id", h.DeleteUserByID)
}
