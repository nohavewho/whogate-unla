package server

import (
	"github.com/gin-gonic/gin"
	"github.com/mcp-ecosystem/mcp-gateway/internal/auth"
	"go.uber.org/zap"
	"net/http"
)

// loggerMiddleware creates a logging middleware
func (s *Server) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		s.logger.Info("incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.Request.RemoteAddr),
		)

		c.Next()

		s.logger.Info("outgoing response",
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
		)
	}
}

// recoveryMiddleware creates a recovery middleware
func (s *Server) recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}

// authMiddleware creates an authentication middleware
func (s *Server) authMiddleware(authenticator auth.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := authenticator.Authenticate(c.Request.Context(), c.Request); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// corsMiddleware creates a CORS middleware
func (s *Server) corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.Next()
			return
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
				c.Header("Access-Control-Allow-Credentials", "true")
				break
			}
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
