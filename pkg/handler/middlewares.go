package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/google/uuid"
)

const (
	requestIDHeaderName = "X-Request-ID"
	contextRequestIDKey = "request_id"
)

//SetRequestIDMiddleware middleware for storing RequestID in Context
func SetRequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeaderName)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Writer.Header().Set(requestIDHeaderName, requestID)
		c.Set(contextRequestIDKey, requestID)

		// before request
		c.Next()
		// after request

	}
}

// Logging middleware
func Logging() gin.HandlerFunc {
	return gin.LoggerWithConfig(
		gin.LoggerConfig{
			SkipPaths: []string{"/api/healthcheck", "/metrics"},
			Formatter: func(param gin.LogFormatterParams) string {
				return fmt.Sprintf("%v method: %s path: %s response_time: %.8f status: %d request_id: %s\n",
					param.TimeStamp.Format(time.RFC3339),
					param.Method,
					param.Path,
					param.Latency.Seconds(),
					param.StatusCode,
					param.Keys[contextRequestIDKey],
				)
			}})
}

// Recovery middleware
func Recovery(f func(c *gin.Context, err interface{})) gin.HandlerFunc {
	return RecoveryWithWriter(f, gin.DefaultErrorWriter)
}

// RecoveryWithWriter middleware
func RecoveryWithWriter(f func(c *gin.Context, err interface{}), out io.Writer) gin.HandlerFunc {
	var logger *log.Logger
	if out != nil {
		logger = log.New(out, "\n\n\x1b[31m", log.LstdFlags)
	}

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if logger != nil {
					httprequest, _ := httputil.DumpRequest(c.Request, false)
					goErr := errors.Wrap(err, 3)
					reset := string([]byte{27, 91, 48, 109})
					logger.Printf("[Recovery] panic recovered:\n\n%s%s\n\n%s%s", httprequest, goErr.Error(), goErr.Stack(), reset)
				}

				f(c, err)
			}
		}()
		c.Next()
	}
}

// RecoveryHandler handler
func RecoveryHandler(c *gin.Context, err interface{}) {
	c.JSON(http.StatusInternalServerError, errorResponse{errorMessageServerError})
}

// NoRouteHandler handler
func NoRouteHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, errorResponse{errorMessageNotFound})
}

// NoMethodHandler handler
func NoMethodHandler(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, errorResponse{errorMessageMethodNotAllowed})
}
