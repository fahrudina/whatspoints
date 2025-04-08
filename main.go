package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/wa-serv/config"
	"github.com/wa-serv/database"
	"github.com/wa-serv/whatsapp"
)

// Global variables
var db *sql.DB

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

	// Listen for termination signals
	waitForTermination(client)
}

func initializeDatabase() {
	// Get database configuration from loaded environment variables
	dbConfig := config.Env

	// Open MySQL database connection
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dbConfig.DBUsername, dbConfig.DBPassword, dbConfig.DBHost, dbConfig.DBPort, dbConfig.DBName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open MySQL database connection: %v\n", err)
		os.Exit(1)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to ping MySQL database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully connected to MySQL database")

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
	fmt.Println("All tables initialized successfully")
}

func waitForTermination(client *whatsapp.Client) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
	if db != nil {
		db.Close()
	}
}
