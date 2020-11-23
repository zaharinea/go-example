package handler

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthcheck(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	handlers := &Handler{}

	router.GET("/", handlers.Healthcheck)
	w := performRequest(router, "GET", "/", "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{\"status\":\"ok\"}", w.Body.String())
}
