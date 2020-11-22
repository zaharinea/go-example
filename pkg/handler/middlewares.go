package handler

import (
	"github.com/gin-gonic/gin"
	_ "github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeaderName = "X-Request-ID"
const ContextRequestIDKey = "request_id"

func SetRequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeaderName)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Writer.Header().Set(RequestIDHeaderName, requestID)
		c.Set(ContextRequestIDKey, requestID)

		// before request
		c.Next()
		// after request

	}
}
