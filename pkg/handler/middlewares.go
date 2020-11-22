package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//RequestIDHeaderName RequestID Header Name
const RequestIDHeaderName = "X-Request-ID"

//ContextRequestIDKey RequestID in Context
const ContextRequestIDKey = "request_id"

//SetRequestIDMiddleware middleware for storing RequestID in Context
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
