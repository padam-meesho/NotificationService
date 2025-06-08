package middlewares

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string // this is a custom data type we have created.

const traceIDKey contextKey = "trace_id"

func TraceMiddleware() gin.HandlerFunc {
	// this middleware simply generates a new trace id for each incoming request
	// and adds it via a header to our request
	return func(c *gin.Context) {
		traceID := uuid.New().String()
		c.Header("X-Trace-Id", traceID)

		ctx := context.WithValue(c.Request.Context(), traceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// what is the need for the custom data type or the use of ContextKeys?
// understand this, we can have multiple functions which want to write with the same header name to the request
// so what we do is create a custom type for it, how that helps?
// middlewares.contextKey["traceId"]!=logger.contextKey["traceId"]
