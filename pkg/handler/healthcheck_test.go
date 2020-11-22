package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealthcheck(t *testing.T) {
	router := gin.Default()
	handlers := &Handler{}

	router.GET("/", handlers.Healthcheck)
	w := performRequest(router, "GET", "/")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, w.Body.String(), "{\"status\":\"ok\"}")
}
