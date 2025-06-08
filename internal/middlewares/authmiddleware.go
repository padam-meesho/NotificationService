package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthCheck() gin.HandlerFunc {
	// consume the auth header and check if it matches our needs. it should be set to "Bearer password123"
	// only for middleware we shall use gin.HandlerFunc
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.Split(token, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{
				"error": "Invalid authorization format. Expected 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		if parts[1] != "password123" {
			c.JSON(401, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}
		c.Next()
	} // the middleware work is over now, it shall now pass it to the next one.
}
