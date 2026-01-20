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
	addSender := flag.Bool("add-sender", false, "Add a new WhatsApp phone number using QR code")
	addSenderWithCode := flag.String("add-sender-code", "", "Add a new WhatsApp phone number using pairing code (provide phone number with country code, e.g., +1234567890)")
	flag.Parse()

	if *clearSessions {
		if err := whatsapp.ClearAllSessions(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clear sessions: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All WhatsApp sessions cleared successfully")
		os.Exit(0)
	}

	if *addSender {
		addNewSenderWithQR()
		os.Exit(0)
	}

	if *addSenderWithCode != "" {
		addNewSenderWithPairingCode(*addSenderWithCode)
		os.Exit(0)
	}

	// Load environment variables
	config.LoadEnv()
	fmt.Println("Environment variables loaded successfully")

	// Initialize database
	initializeDatabase()
	fmt.Println("Database initialized successfully")

	// Initialize WhatsApp ClientManager with multi-sender support
	connectionString := database.BuildPostgresConnectionString()
	clientManager, err := whatsapp.NewClientManager(db, connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize ClientManager: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("WhatsApp ClientManager initialized successfully")

	// Start API server with ClientManager
	startAPIServerWithClientManager(clientManager)

	// Listen for termination signals
	waitForTerminationWithClientManager(clientManager)
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

	// Note: Whatsmeow session storage tables are automatically initialized by sqlstore.New()
	// in the ClientManager, so we don't need to manually create them here
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
		fmt.Printf("  GET  /api/senders      - List available senders\n")
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

func startAPIServerWithClientManager(clientManager *whatsapp.ClientManager) {
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

	// Create API server with multi-client support
	apiServer := api.NewAPIServerWithClientManager(db, clientManager, username, password, port)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting API server on port %s...\n", port)
		fmt.Printf("API endpoints:\n")
		fmt.Printf("  POST /api/send-message - Send WhatsApp message\n")
		fmt.Printf("  GET  /api/status       - Get service status\n")
		fmt.Printf("  GET  /health           - Health check\n")
		fmt.Printf("  GET  /api/senders      - List available senders\n")
		fmt.Printf("Basic Auth: %s / %s\n", username, "***")

		// List available senders
		senders := clientManager.ListClients()
		if len(senders) > 0 {
			fmt.Printf("\nAvailable senders: %d\n", len(senders))
			for i, senderID := range senders {
				fmt.Printf("  %d. %s\n", i+1, senderID)
			}
		} else {
			fmt.Println("\n⚠ No senders available. Add a sender using:")
			fmt.Println("  ./whatspoints -add-sender")
			fmt.Println("  ./whatspoints -add-sender-code=+PHONE_NUMBER")
		}

		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start API server: %v", err)
		}
	}()

	httpServer = apiServer.GetHTTPServer()
}

func waitForTerminationWithClientManager(clientManager *whatsapp.ClientManager) {
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

	// Disconnect all WhatsApp clients
	if clientManager != nil {
		clientManager.DisconnectAll()
		fmt.Println("All WhatsApp clients disconnected")
	}

	// Close database connection
	if db != nil {
		db.Close()
		fmt.Println("Database connection closed")
	}

	fmt.Println("Shutdown complete")
}

// addNewSenderWithQR adds a new WhatsApp phone number using QR code scanning
func addNewSenderWithQR() {
	setupAndAddSender(func(clientManager *whatsapp.ClientManager) error {
		fmt.Println("\n=== QR Code Pairing Method ===")
		fmt.Println("This will display a QR code for scanning with WhatsApp")
		fmt.Println()

		_, err := clientManager.AddNewClient()
		return err
	})
}

// addNewSenderWithPairingCode adds a new WhatsApp phone number using SMS pairing code
func addNewSenderWithPairingCode(phoneNumber string) {
	// Validate phone number format
	if len(phoneNumber) < 10 {
		fmt.Fprintf(os.Stderr, "Invalid phone number. Please provide with country code (e.g., +1234567890)\n")
		os.Exit(1)
	}

	// Clean phone number
	cleanedPhone := cleanPhoneNumber(phoneNumber)

	setupAndAddSender(func(clientManager *whatsapp.ClientManager) error {
		fmt.Println("\n=== Phone Number Pairing Method ===")
		fmt.Println("This will send a pairing code via SMS to the phone number")
		fmt.Println()

		_, err := clientManager.AddNewClientWithPairingCode(cleanedPhone)
		return err
	})
}

// setupAndAddSender handles common setup and calls the provided add function
func setupAndAddSender(addFunc func(*whatsapp.ClientManager) error) {
	// Load environment variables
	config.LoadEnv()
	fmt.Println("Environment variables loaded successfully")

	// Initialize database
	connectionString := database.BuildPostgresConnectionString()

	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	// Initialize required tables
	if err := database.InitSendersTable(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize senders table: %v\n", err)
		os.Exit(1)
	}

	// Note: Whatsmeow tables are automatically created by sqlstore.New() in ClientManager

	// Create ClientManager
	clientManager, err := whatsapp.NewClientManager(db, connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client manager: %v\n", err)
		os.Exit(1)
	}

	// Call the add function
	if err := addFunc(clientManager); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add new sender: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ New WhatsApp phone number added successfully!")
	fmt.Println("You can now use this sender with the API.")

	// List all available senders
	senders := clientManager.ListClients()
	fmt.Printf("\nTotal senders available: %d\n", len(senders))
	for i, senderID := range senders {
		fmt.Printf("%d. Sender ID: %s\n", i+1, senderID)
	}
}

// cleanPhoneNumber removes +, spaces, and other non-digit characters
func cleanPhoneNumber(phone string) string {
	cleaned := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			cleaned += string(char)
		}
	}
	return cleaned
}
