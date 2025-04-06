package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/mdp/qrterminal/v3"
	"github.com/wa-serv/handlers"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Client struct {
	whatsmeowClient *whatsmeow.Client
}

func InitializeWhatsAppClient(db *sql.DB) *Client {
	// Set up database connection for storing session data
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", "file:whatsappdata.db?_foreign_keys=on", dbLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to SQLite database: %v\n", err)
		os.Exit(1)
	}

	deviceStore, err := container.GetFirstDevice()
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
