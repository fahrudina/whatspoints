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
	"github.com/wa-serv/repository"
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

// HandleEvent processes WhatsApp events (exported for use in other packages)
func HandleEvent(evt interface{}, db *sql.DB, client *whatsmeow.Client) {
	switch v := evt.(type) {
	case *events.Message:
		handlers.HandleMessageEvent(v, db, client)
	case *events.Connected:
		handleConnected(client)
	case *events.Disconnected:
		handleDisconnected(client)
	case *events.PairSuccess:
		fmt.Println("Successfully paired with device")
	case *events.LoggedOut:
		handleLogout(v, db, client)
	case *events.StreamReplaced:
		handleStreamReplaced(client)
	case *events.StreamError:
		handleStreamError(v, client)
	}
}

// handleConnected handles connection events
func handleConnected(client *whatsmeow.Client) {
	if client.Store.ID != nil {
		senderID := client.Store.ID.User
		log.Printf("✓ Client %s connected to WhatsApp", senderID)
	} else {
		fmt.Println("✓ Connected to WhatsApp")
	}
}

// handleDisconnected handles disconnection events
func handleDisconnected(client *whatsmeow.Client) {
	if client.Store.ID != nil {
		senderID := client.Store.ID.User
		log.Printf("Client %s disconnected - whatsmeow handles automatic reconnection internally", senderID)
	} else {
		fmt.Println("Disconnected from WhatsApp - whatsmeow handles automatic reconnection internally")
	}
	// IMPORTANT: Do NOT manually reconnect here
	// Whatsmeow has built-in reconnection logic
	// Manual reconnection attempts can trigger WhatsApp's security system
	// which causes "unexpected issue" logouts
}

// handleStreamReplaced handles stream replacement events
func handleStreamReplaced(client *whatsmeow.Client) {
	if client.Store.ID != nil {
		senderID := client.Store.ID.User
		log.Printf("⚠ Client %s - stream replaced by another session", senderID)
	} else {
		fmt.Println("⚠ Stream replaced - this connection was replaced by another session")
	}
}

// handleStreamError handles stream error events
func handleStreamError(evt *events.StreamError, client *whatsmeow.Client) {
	if client.Store.ID != nil {
		senderID := client.Store.ID.User
		log.Printf("⚠ Client %s - stream error (code: %s) - automatic reconnect will handle it", senderID, evt.Code)
	} else {
		log.Printf("⚠ Stream error (code: %s) - automatic reconnect will handle it", evt.Code)
	}

	// Stream errors (like 503) are typically handled by automatic reconnection
	// Only log for monitoring purposes - the client will attempt to reconnect
}

// handleLogout handles the LoggedOut event
func handleLogout(evt *events.LoggedOut, db *sql.DB, client *whatsmeow.Client) {
	reason := evt.Reason
	fmt.Printf("Device logged out - Reason: %d (%s)\n", reason, reason.String())

	if client.Store.ID == nil {
		fmt.Println("Warning: Client has no ID, cannot update sender status")
		return
	}

	senderID := client.Store.ID.User

	// For ANY logout event from WhatsApp, mark as inactive
	// Do NOT try to reconnect - WhatsApp security system may have triggered this
	// Reconnection attempts can cause more security flags
	fmt.Printf("WhatsApp logged out device %s - marking as inactive\n", senderID)

	// Update sender status to inactive
	if err := repository.UpdateSenderStatus(db, senderID, false); err != nil {
		log.Printf("Failed to update sender status for %s: %v", senderID, err)
	} else {
		fmt.Printf("Sender %s marked as inactive\n", senderID)
	}

	fmt.Printf("⚠ To reconnect sender %s, please re-register via QR code or pairing code\n", senderID)
} // handleEvent is kept for backward compatibility within this package
func handleEvent(evt interface{}, db *sql.DB, client *whatsmeow.Client) {
	HandleEvent(evt, db, client)
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
