package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wa-serv/internal/application"
	"github.com/wa-serv/internal/infrastructure"
	"github.com/wa-serv/internal/presentation"
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

// Start starts the API server
func (s *APIServer) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the API server
func (s *APIServer) Shutdown() error {
	return s.httpServer.Close()
}

// GetRouter returns the gin router for testing
func (s *APIServer) GetRouter() *gin.Engine {
	return s.router
}
