package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseHealthcheck struct {
	Status string `json:"status"`
}

// Healthcheck handler
// @Summary Healthcheck
// @Tags healthcheck
// @Produce  json
// @Success 200 {object} ResponseHealthcheck
// @Router /api/healthcheck [get]
func (h *Handler) Healthcheck(c *gin.Context) {
	c.JSON(http.StatusOK, ResponseHealthcheck{Status: "ok"})
}
