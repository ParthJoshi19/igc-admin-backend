package middleware

import (
	"fmt"
	"log"
	"time"
	"github.com/gin-gonic/gin"
)

// Logger is a middleware that logs HTTP requests
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// CORS middleware - backup implementation (currently using gin-contrib/cors in main.go)
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Allow specific origins
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		}

		isAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		}

		// Handle preflight request
		if c.Request.Method == "OPTIONS" {
			if isAllowed {
				c.AbortWithStatus(204)
			} else {
				c.AbortWithStatus(403)
			}
			return
		}

		c.Next()
	}
}


// ErrorHandler is a middleware that handles panics and errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(500, gin.H{
					"error":   "Internal server error",
					"message": "Something went wrong on our end",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// BasicAuth is a simple authentication middleware (for admin routes)
// In production, you should use JWT or proper session management
func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For now, we'll skip authentication
		// In a real application, you would validate JWT tokens here
		
		// Example of how you might check for a simple API key:
		// apiKey := c.GetHeader("X-API-Key")
		// if apiKey != "your-secret-api-key" {
		//     c.JSON(401, gin.H{"error": "Unauthorized"})
		//     c.Abort()
		//     return
		// }
		
		c.Next()
	}
}