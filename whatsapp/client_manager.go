package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// getLogLevel returns the WhatsApp log level from environment or default to INFO
func getLogLevel() string {
	logLevel := os.Getenv("WHATSAPP_LOG_LEVEL")
	if logLevel == "" {
		return "INFO"
	}
	return logLevel
}

// ClientManager manages multiple WhatsApp clients
type ClientManager struct {
	db              *sql.DB
	container       *sqlstore.Container
	clients         map[string]*whatsmeow.Client // key: sender_id
	defaultSenderID string
	mu              sync.RWMutex
}

// NewClientManager creates a new client manager
func NewClientManager(db *sql.DB, connectionString string) (*ClientManager, error) {
	dbLog := waLog.Stdout("Database", getLogLevel(), true)
	container, err := sqlstore.New(context.Background(), "postgres", connectionString, dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database for WhatsApp sessions: %w", err)
	}

	cm := &ClientManager{
		db:        db,
		container: container,
		clients:   make(map[string]*whatsmeow.Client),
	}

	// Initialize with existing devices
	if err := cm.loadExistingClients(); err != nil {
		return nil, fmt.Errorf("failed to load existing clients: %w", err)
	}

	return cm, nil
}

// loadExistingClients loads all existing WhatsApp clients from the database
func (cm *ClientManager) loadExistingClients() error {
	devices, err := cm.container.GetAllDevices(context.Background())
	if err != nil {
		return err
	}

	logLevel := getLogLevel()

	for _, device := range devices {
		if device.ID != nil {
			// Get or create sender record
			senderID := device.ID.User
			cm.ensureSenderRecord(senderID, device.ID.User)

			// Create client
			clientLog := waLog.Stdout(fmt.Sprintf("Client-%s", senderID), logLevel, true)
			client := whatsmeow.NewClient(device, clientLog)

			// Add event handler
			client.AddEventHandler(func(evt interface{}) {
				handleEvent(evt, cm.db, client)
			})

			// Connect the client
			if err := client.Connect(); err != nil {
				log.Printf("Failed to connect client %s: %v", senderID, err)
				continue
			}

			cm.mu.Lock()
			cm.clients[senderID] = client

			// Set as default if it's the first one
			if cm.defaultSenderID == "" {
				cm.defaultSenderID = senderID
			}
			cm.mu.Unlock()
		}
	}

	return nil
}

// ensureSenderRecord ensures a sender record exists in the database
func (cm *ClientManager) ensureSenderRecord(senderID, phoneNumber string) {
	query := `
		INSERT INTO senders (sender_id, phone_number, name, is_default, is_active)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (sender_id) DO NOTHING
	`
	cm.mu.RLock()
	isDefault := cm.defaultSenderID == ""
	cm.mu.RUnlock()

	_, err := cm.db.Exec(query, senderID, phoneNumber, fmt.Sprintf("Sender %s", senderID), isDefault, true)
	if err != nil {
		log.Printf("Failed to create sender record: %v", err)
	}
}

// GetClient returns a specific client by sender ID
func (cm *ClientManager) GetClient(senderID string) (*whatsmeow.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.clients[senderID]
	if !exists {
		return nil, fmt.Errorf("client not found for sender: %s", senderID)
	}
	return client, nil
}

// GetDefaultClient returns the default client
func (cm *ClientManager) GetDefaultClient() (*whatsmeow.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.defaultSenderID == "" {
		// Try to get first device
		devices, err := cm.container.GetAllDevices(context.Background())
		if err != nil || len(devices) == 0 {
			return nil, fmt.Errorf("no default client available")
		}
		// Use first device
		if devices[0].ID != nil {
			return cm.GetClient(devices[0].ID.User)
		}
		return nil, fmt.Errorf("no default client available")
	}

	return cm.GetClient(cm.defaultSenderID)
}

// ListClients returns all available client IDs
func (cm *ClientManager) ListClients() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ids := make([]string, 0, len(cm.clients))
	for id := range cm.clients {
		ids = append(ids, id)
	}
	return ids
}

// DisconnectAll disconnects all clients
func (cm *ClientManager) DisconnectAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, client := range cm.clients {
		client.Disconnect()
	}
}
