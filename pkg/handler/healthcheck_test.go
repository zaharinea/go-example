package handler

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type HealthcheckSuite struct {
	suite.Suite
	router   *gin.Engine
	handlers *Handler
}

func (s *HealthcheckSuite) SetupSuite() {
	gin.SetMode(gin.ReleaseMode)
	s.router = gin.New()
	s.handlers = &Handler{}
}

func (s *HealthcheckSuite) TestHealthcheck() {
	s.router.GET("/", s.handlers.Healthcheck)
	w := performRequest(s.router, "GET", "/", "")
	s.Require().Equal(http.StatusOK, w.Code)
	s.Require().Equal("{\"status\":\"ok\"}", w.Body.String())
}

func TestHealthcheckSuite(t *testing.T) {
	suite.Run(t, new(HealthcheckSuite))
}
