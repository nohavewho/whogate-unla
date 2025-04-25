package core

import (
	"github.com/mcp-ecosystem/mcp-gateway/internal/common/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mcp-ecosystem/mcp-gateway/internal/auth"
	"go.uber.org/zap"
)

// loggerMiddleware logs incoming requests and outgoing responses
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

// recoveryMiddleware recovers from panics and returns 500 error
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

// authMiddleware authenticates incoming requests
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

// corsMiddleware handles CORS configuration
func (s *Server) corsMiddleware(cors *config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.Next()
			return
		}

		allowed := false
		for _, allowedOrigin := range cors.AllowOrigins {
			if allowedOrigin == "*" || origin == allowedOrigin {
				allowed = true
				c.Header("Access-Control-Allow-Origin", allowedOrigin)
				break
			}
		}

		if !allowed {
			c.Next()
			return
		}

		if len(cors.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(cors.AllowMethods, ", "))
		}

		if len(cors.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(cors.AllowHeaders, ", "))
		}

		if len(cors.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(cors.ExposeHeaders, ", "))
		}

		if cors.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
