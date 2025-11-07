package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver for Supabase
	"github.com/wa-serv/api"
	"github.com/wa-serv/config"
	"github.com/wa-serv/database"
	"github.com/wa-serv/whatsapp"
)

// Global variables
var db *sql.DB
var httpServer *http.Server

func main() {

	clearSessions := flag.Bool("clear-sessions", false, "Clear all WhatsApp sessions")
	flag.Parse()

	if *clearSessions {
		if err := whatsapp.ClearAllSessions(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clear sessions: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All WhatsApp sessions cleared successfully")
		os.Exit(0)
	}

	// Load environment variables
	config.LoadEnv()
	fmt.Println("Environment variables loaded successfully")

	// Initialize database
	initializeDatabase()
	fmt.Println("Database initialized successfully")

	// Initialize WhatsApp client
	client := whatsapp.InitializeWhatsAppClient(db)
	fmt.Println("WhatsApp client initialized successfully")

	// Start API server
	startAPIServer(client)

	// Listen for termination signals
	waitForTermination(client)
}

func initializeDatabase() {
	// Supabase Transaction Pooler connection string
	connectionString := database.BuildPostgresConnectionString()

	fmt.Printf("Connecting to Supabase Transaction Pooler...\n")

	// Connect directly with sql.DB
	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to Supabase (Postgres) database: %v\n", err)
		os.Exit(1)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to ping Supabase (Postgres) database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully connected to Supabase Transaction Pooler")

	// Test the sql.DB connection
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to ping SQL database: %v\n", err)
		os.Exit(1)
	}

	// Initialize tables
	if err := database.InitMemberTable(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize member table: %v\n", err)
		os.Exit(1)
	}
	if err := database.InitImageTable(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize images table: %v\n", err)
		os.Exit(1)
	}
	if err := database.InitPointsTable(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize points table: %v\n", err)
		os.Exit(1)
	}
	if err := database.InitReceiptsTable(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize receipts table: %v\n", err)
		os.Exit(1)
	}
	if err := database.InitPointTransactionsTable(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize transactions table: %v\n", err)
		os.Exit(1)
	}
	if err := database.InitItemsTable(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize items table: %v\n", err)
		os.Exit(1)
	}
	if err := database.InitOrdersTable(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize orders table: %v\n", err)
		os.Exit(1)
	}
	if err := database.InitOrderItemsTable(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize order_items table: %v\n", err)
		os.Exit(1)
	}

	// Initialize senders table for multi-sender support
	if err := database.InitSendersTable(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize senders table: %v\n", err)
		os.Exit(1)
	}

	// Initialize Whatsmeow session storage tables
	if err := database.InitWhatsmeowTables(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize Whatsmeow tables: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Whatsmeow session storage tables initialized successfully")
	fmt.Println("All tables initialized successfully")
}

func startAPIServer(client *whatsapp.Client) {
	// Get API configuration from environment variables
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080" // Default port
	}

	username := os.Getenv("API_USERNAME")
	if username == "" {
		username = "admin" // Default username
	}

	password := os.Getenv("API_PASSWORD")
	if password == "" {
		log.Fatal("API_PASSWORD environment variable is required")
	}

	// Create API server using clean architecture
	apiServer := api.NewAPIServer(db, client.GetWhatsmeowClient(), username, password, port)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting API server on port %s...\n", port)
		fmt.Printf("API endpoints:\n")
		fmt.Printf("  POST /api/send-message - Send WhatsApp message\n")
		fmt.Printf("  GET  /api/status       - Get service status\n")
		fmt.Printf("  GET  /health           - Health check\n")
		fmt.Printf("Basic Auth: %s / %s\n", username, "***")

		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	// Store reference for graceful shutdown
	httpServer = &http.Server{}
}

func waitForTermination(client *whatsapp.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\nShutting down gracefully...")

	// Shutdown API server
	if httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown API server: %v", err)
		} else {
			fmt.Println("API server stopped")
		}
	}

	// Disconnect WhatsApp client
	if client != nil {
		client.Disconnect()
		fmt.Println("WhatsApp client disconnected")
	}

	// Close database connection
	if db != nil {
		db.Close()
		fmt.Println("Database connection closed")
	}

	fmt.Println("Shutdown complete")
}
