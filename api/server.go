package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wa-serv/internal/application"
	"github.com/wa-serv/internal/infrastructure"
	"github.com/wa-serv/internal/presentation"
	"github.com/wa-serv/whatsapp"
	"go.mau.fi/whatsmeow"
)

// APIServer represents the API server using clean architecture
type APIServer struct {
	router     *gin.Engine
	httpServer *http.Server
}

// NewAPIServer creates a new API server instance using clean architecture
func NewAPIServer(db *sql.DB, client *whatsmeow.Client, username, password string, port string) *APIServer {
	// Infrastructure layer - use repository with database support
	whatsappRepo := infrastructure.NewWhatsAppRepositoryWithDB(client, db)

	// Application layer
	messageService := application.NewMessageService(whatsappRepo)
	authService := application.NewAuthService(username, password)

	// Presentation layer
	messageHandler := presentation.NewMessageHandler(messageService, authService)
	router := presentation.NewRouter(messageHandler, authService)

	// Setup routes
	ginRouter := router.SetupRoutes()

	// Configure HTTP server
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      ginRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &APIServer{
		router:     ginRouter,
		httpServer: httpServer,
	}
}

// NewAPIServerWithClientManager creates a new API server with multi-client support
func NewAPIServerWithClientManager(db *sql.DB, clientManager *whatsapp.ClientManager, username, password string, port string) *APIServer {
	// Infrastructure layer - use repository with client manager for dynamic client updates
	whatsappRepo := infrastructure.NewWhatsAppRepositoryWithClientManager(db, clientManager)

	// Application layer
	messageService := application.NewMessageService(whatsappRepo)
	authService := application.NewAuthService(username, password)
	registrationService := application.NewSenderRegistrationService(db, clientManager)

	// Presentation layer
	messageHandler := presentation.NewMessageHandler(messageService, authService)
	registrationHandler := presentation.NewSenderRegistrationHandler(registrationService, authService)
	router := presentation.NewRouterWithRegistration(messageHandler, registrationHandler, authService)

	// Setup routes
	ginRouter := router.SetupRoutes()

	// Configure HTTP server
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      ginRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &APIServer{
		router:     ginRouter,
		httpServer: httpServer,
	}
}

// Start starts the API server
func (s *APIServer) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the API server
func (s *APIServer) Shutdown() error {
	return s.httpServer.Close()
}

// GetHTTPServer returns the underlying HTTP server
func (s *APIServer) GetHTTPServer() *http.Server {
	return s.httpServer
}

// GetRouter returns the gin router for testing
func (s *APIServer) GetRouter() *gin.Engine {
	return s.router
}
