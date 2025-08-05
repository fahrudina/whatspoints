package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver for Supabase
	"github.com/mdp/qrterminal/v3"
	"github.com/wa-serv/database"
	"github.com/wa-serv/handlers"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Client struct {
	whatsmeowClient *whatsmeow.Client
}

// GetWhatsmeowClient returns the underlying whatsmeow client
func (c *Client) GetWhatsmeowClient() *whatsmeow.Client {
	return c.whatsmeowClient
}

func InitializeWhatsAppClient(db *sql.DB) *Client {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Build Supabase PostgreSQL connection string for WhatsApp session storage
	connectionString := database.BuildPostgresConnectionString()

	fmt.Printf("Connecting WhatsApp client to Supabase PostgreSQL...\n")

	// Set up database connection for storing WhatsApp session data using Supabase PostgreSQL
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New(context.Background(), "postgres", connectionString, dbLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to Supabase PostgreSQL database for WhatsApp sessions: %v\n", err)
		os.Exit(1)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get device: %v\n", err)
		os.Exit(1)
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	whatsmeowClient := whatsmeow.NewClient(deviceStore, clientLog)
	whatsmeowClient.AddEventHandler(func(evt interface{}) {
		handleEvent(evt, db, whatsmeowClient)
	})

	// Connect to WhatsApp
	connectToWhatsApp(whatsmeowClient)

	return &Client{whatsmeowClient: whatsmeowClient}
}

func connectToWhatsApp(client *whatsmeow.Client) {
	if client.Store.ID == nil {
		// No ID stored, needs QR code login
		qrChan, _ := client.GetQRChannel(context.Background())
		err := client.Connect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
			os.Exit(1)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in
		err := client.Connect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
			os.Exit(1)
		}
	}
}

func handleEvent(evt interface{}, db *sql.DB, client *whatsmeow.Client) {
	switch v := evt.(type) {
	case *events.Message:
		handlers.HandleMessageEvent(v, db, client)
	case *events.Connected:
		fmt.Println("Connected to WhatsApp")
	case *events.Disconnected:
		fmt.Println("Disconnected from WhatsApp")
	case *events.PairSuccess:
		fmt.Println("Successfully paired with device")
	case *events.LoggedOut:
		fmt.Println("Device logged out")
	}
}

func (c *Client) Disconnect() {
	c.whatsmeowClient.Disconnect()
}

func ClearAllSessions() error {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}

	// Build Supabase PostgreSQL connection string
	connectionString := database.BuildPostgresConnectionString()

	// Connect to the same Supabase PostgreSQL database
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New(context.Background(), "postgres", connectionString, dbLog)
	if err != nil {
		return fmt.Errorf("failed to connect to Supabase PostgreSQL database: %v", err)
	}

	// Get all devices
	devices, err := container.GetAllDevices(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get devices: %v", err)
	}

	// Delete each device
	for _, device := range devices {
		if err := container.DeleteDevice(context.Background(), device); err != nil {
			return fmt.Errorf("failed to delete device %s: %v", device.ID, err)
		}
	}

	return nil
}
