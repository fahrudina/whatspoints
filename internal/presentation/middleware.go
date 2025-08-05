package presentation

import (
	"github.com/gin-gonic/gin"
	"github.com/wa-serv/internal/domain"
)

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
