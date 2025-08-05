package presentation

import (
	"github.com/gin-gonic/gin"
	"github.com/wa-serv/internal/domain"
)

// BasicAuthMiddleware creates a basic auth middleware
func BasicAuthMiddleware(authService domain.AuthService) gin.HandlerFunc {
	return gin.BasicAuthForRealm(gin.Accounts{}, "WhatsPoints API")
}

// CustomBasicAuth creates a custom basic auth middleware using the auth service
func CustomBasicAuth(authService domain.AuthService) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		// This will be dynamically validated in the auth service
		// For now, we'll use a placeholder and override the validation
	})
}

// AuthMiddleware validates credentials using the auth service
func AuthMiddleware(authService domain.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, hasAuth := c.Request.BasicAuth()

		if !hasAuth || !authService.ValidateCredentials(username, password) {
			c.Header("WWW-Authenticate", `Basic realm="WhatsPoints API"`)
			c.AbortWithStatus(401)
			return
		}

		c.Next()
	}
}
