package routes

import (
	"AuthService/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, authHandler *handlers.AuthHandler) {

	r.Use(func(c *gin.Context) {
		c.Set("client_ip", c.ClientIP())
		c.Next()
	})

	auth := r.Group("/auth")
	{
		auth.POST("/tokens", authHandler.CreateTokens)
		auth.POST("/refresh", authHandler.RefreshTokens)
	}
}
