package presentation

import (
	"github.com/gin-gonic/gin"
	"github.com/wa-serv/internal/domain"
)

type Router struct {
	messageHandler *MessageHandler
	authService    domain.AuthService
}

// NewRouter creates a new router
func NewRouter(messageHandler *MessageHandler, authService domain.AuthService) *Router {
	return &Router{
		messageHandler: messageHandler,
		authService:    authService,
	}
}

// SetupRoutes sets up all the routes
func (r *Router) SetupRoutes() *gin.Engine {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint (no auth required)
	router.GET("/health", r.messageHandler.HealthCheck)

	// API routes with Basic Auth
	apiRoutes := router.Group("/api")
	apiRoutes.Use(AuthMiddleware(r.authService))
	{
		apiRoutes.POST("/send-message", r.messageHandler.SendMessage)
		apiRoutes.GET("/status", r.messageHandler.GetStatus)
	}

	return router
}
