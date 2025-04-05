package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/mattn/go-sqlite3"    // SQLite driver
	"github.com/wa-serv/config"
	"github.com/wa-serv/database"
	"github.com/wa-serv/processor"
)

// Add a global database variable
var db *sql.DB

func main() {
	// Set up database connection for storing session data
	dbLog := waLog.Stdout("Database", "DEBUG", true)

	// SQLite connection string - this will create a file named whatsappdata.db
	container, err := sqlstore.New("sqlite3", "file:whatsappdata.db?_foreign_keys=on", dbLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get device: %v\n", err)
		os.Exit(1)
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Make the client globally accessible for the reply function
	globalClient = client

	// Set up event handler
	client.AddEventHandler(eventHandler)

	// Connect to WhatsApp
	if client.Store.ID == nil {
		// No ID stored, needs QR code login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
			os.Exit(1)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Print QR code to terminal
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in
		err = client.Connect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
			os.Exit(1)
		}
	}

	// Get database configuration
	dbConfig := config.GetDatabaseConfig()

	// Open MySQL database connection using the config
	db, err = sql.Open(dbConfig.Driver, dbConfig.BuildConnectionString())
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
	defer db.Close()

	// Initialize the member table
	if err := database.InitMemberTable(db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize member table: %v\n", err)
		os.Exit(1)
	}

	// Listen for termination signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}

// Global client for use in event handlers
var globalClient *whatsmeow.Client

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		// Access the message content
		var msgText string
		if v.Message.GetExtendedTextMessage().GetText() != "" {
			msgText = v.Message.GetExtendedTextMessage().GetText()
		} else {
			msgText = v.Message.GetConversation()
		}

		fmt.Printf("Received message from %s: %s\n", v.Info.Sender.String(), msgText)

		// Process registration commands
		err := processor.ProcessRegistration(globalClient, db, msgText, v.Info.Sender.String())
		if err != nil {
			fmt.Printf("Registration processing error: %v\n", err)
		}

		// Example of how to reply to a message
		if msgText == "ping" {
			replyToMessage(v)
		} else if msgText == "help" {
			sendHelpMessage(v)
		}
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

func replyToMessage(evt *events.Message) {
	// Create a message using the correct proto structure
	msg := &waProto.Message{
		Conversation: proto.String("pong"),
	}

	// Send the message with the correct method signature
	_, err := globalClient.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}

func sendHelpMessage(evt *events.Message) {
	// Create a help message
	helpText := `Available commands:
- ping: Bot responds with "pong"
- help: Shows this help message`

	msg := &waProto.Message{
		Conversation: proto.String(helpText),
	}

	// Send the message with the correct method signature
	_, err := globalClient.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Error sending help message: %v\n", err)
	}
}
