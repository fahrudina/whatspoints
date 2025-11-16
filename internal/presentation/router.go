package presentation

import (
	"github.com/gin-gonic/gin"
	"github.com/wa-serv/internal/domain"
)

type Router struct {
	messageHandler             *MessageHandler
	senderRegistrationHandler  *SenderRegistrationHandler
	authService                domain.AuthService
}

// NewRouter creates a new router
func NewRouter(messageHandler *MessageHandler, authService domain.AuthService) *Router {
	return &Router{
		messageHandler: messageHandler,
		authService:    authService,
	}
}

// NewRouterWithRegistration creates a new router with sender registration support
func NewRouterWithRegistration(messageHandler *MessageHandler, senderRegistrationHandler *SenderRegistrationHandler, authService domain.AuthService) *Router {
	return &Router{
		messageHandler:            messageHandler,
		senderRegistrationHandler: senderRegistrationHandler,
		authService:               authService,
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

	// Serve static files for the web UI (no auth required)
	router.StaticFile("/", "./web/index.html")
	router.StaticFile("/register", "./web/register.html")
	router.Static("/web", "./web")

	// API routes with Basic Auth
	apiRoutes := router.Group("/api")
	apiRoutes.Use(AuthMiddleware(r.authService))
	{
		apiRoutes.POST("/send-message", r.messageHandler.SendMessage)
		apiRoutes.GET("/status", r.messageHandler.GetStatus)
		apiRoutes.GET("/senders", r.messageHandler.ListSenders)

		// Sender registration endpoints (if handler is available)
		if r.senderRegistrationHandler != nil {
			apiRoutes.POST("/register-sender-qr", r.senderRegistrationHandler.StartQRRegistration)
			apiRoutes.POST("/register-sender-code", r.senderRegistrationHandler.StartCodeRegistration)
			apiRoutes.GET("/register-sender-status/:sessionId", r.senderRegistrationHandler.GetRegistrationStatus)
		}
	}

	// Fallback for SPA routing
	router.NoRoute(func(c *gin.Context) {
		c.File("./web/index.html")
	})

	return router
}
