package presentation

import (
	"fmt"
	"os"
	"path/filepath"

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

	// Determine web directory path
	webDir := r.findWebDirectory()
	fmt.Printf("Using web directory: %s\n", webDir)

	// Serve static files for the web UI (no auth required)
	indexPath := filepath.Join(webDir, "index.html")
	registerPath := filepath.Join(webDir, "register.html")

	router.StaticFile("/", indexPath)
	router.StaticFile("/register", registerPath)
	router.Static("/web", webDir)

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
		c.File(indexPath)
	})

	return router
}

// findWebDirectory finds the web directory path, checking multiple possible locations
func (r *Router) findWebDirectory() string {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Warning: Could not get current working directory: %v\n", err)
		return "./web"
	}

	fmt.Printf("Current working directory: %s\n", cwd)

	// Possible locations for web directory
	possiblePaths := []string{
		"./web",                              // Relative to current directory
		filepath.Join(cwd, "web"),            // Absolute path from cwd
		"/app/web",                           // Common Docker/deployment path
		filepath.Join(cwd, "..", "web"),      // One level up
	}

	// Check each possible path
	for _, path := range possiblePaths {
		indexPath := filepath.Join(path, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			fmt.Printf("Found web directory at: %s\n", path)
			return path
		}
	}

	// Default fallback
	fmt.Printf("Warning: Could not find web directory in any expected location. Using default: ./web\n")
	fmt.Printf("Checked paths: %v\n", possiblePaths)
	return "./web"
}
