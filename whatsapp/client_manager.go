package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/wa-serv/repository"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// GetLogLevel returns the WhatsApp log level from environment or default to INFO
func GetLogLevel() string {
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
	reconnecting    map[string]bool // track which clients are currently reconnecting
}

// NewClientManager creates a new client manager
func NewClientManager(db *sql.DB, connectionString string) (*ClientManager, error) {
	dbLog := waLog.Stdout("Database", GetLogLevel(), true)
	container, err := sqlstore.New(context.Background(), "postgres", connectionString, dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database for WhatsApp sessions: %w", err)
	}

	cm := &ClientManager{
		db:           db,
		container:    container,
		clients:      make(map[string]*whatsmeow.Client),
		reconnecting: make(map[string]bool),
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

	logLevel := GetLogLevel()

	// Load default sender from database
	defaultSender, err := repository.GetDefaultSender(cm.db)
	if err == nil && defaultSender != nil {
		cm.defaultSenderID = defaultSender.SenderID
		log.Printf("Loaded default sender from database: %s", cm.defaultSenderID)
	}

	for _, device := range devices {
		if device.ID != nil {
			// Get or create sender record
			senderID := device.ID.User
			cm.ensureSenderRecord(senderID, device.ID.User)

			// Create client
			clientLog := waLog.Stdout(fmt.Sprintf("Client-%s", senderID), logLevel, true)
			client := whatsmeow.NewClient(device, clientLog)

			// Add event handler with client manager awareness
			client.AddEventHandler(func(evt interface{}) {
				cm.handleEventWithCleanup(evt, client)
			})

			// Connect the client
			if err := client.Connect(); err != nil {
				log.Printf("Failed to connect client %s: %v", senderID, err)
				continue
			}

			cm.mu.Lock()
			cm.clients[senderID] = client

			// Set as default if it's the first one and no default was loaded from DB
			if cm.defaultSenderID == "" {
				cm.defaultSenderID = senderID
				// Update database to reflect this
				repository.SetDefaultSender(cm.db, senderID)
			}
			cm.mu.Unlock()
		}
	}

	return nil
}

// ensureSenderRecord ensures a sender record exists in the database
func (cm *ClientManager) ensureSenderRecord(senderID, phoneNumber string) {
	cm.mu.RLock()
	isDefault := cm.defaultSenderID == ""
	cm.mu.RUnlock()

	name := fmt.Sprintf("Sender %s", senderID)

	err := repository.CreateSenderIfNotExists(cm.db, senderID, phoneNumber, name, isDefault)
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

// GetAllClients returns a copy of all clients as a map
func (cm *ClientManager) GetAllClients() map[string]*whatsmeow.Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clientsCopy := make(map[string]*whatsmeow.Client, len(cm.clients))
	for id, client := range cm.clients {
		clientsCopy[id] = client
	}
	return clientsCopy
}

// DisconnectAll disconnects all clients
func (cm *ClientManager) DisconnectAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, client := range cm.clients {
		client.Disconnect()
	}
}

// AddExistingClient adds an already connected client to the manager
func (cm *ClientManager) AddExistingClient(client *whatsmeow.Client, senderID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Add to clients map
	cm.clients[senderID] = client

	// Set as default if it's the first one
	if cm.defaultSenderID == "" {
		cm.defaultSenderID = senderID
	}

	// Add event handler for ongoing message handling with cleanup
	client.AddEventHandler(func(evt interface{}) {
		cm.handleEventWithCleanup(evt, client)
	})
}

// handleEventWithCleanup handles events and performs cleanup for logout events
func (cm *ClientManager) handleEventWithCleanup(evt interface{}, client *whatsmeow.Client) {
	// Handle connected events - mark sender as active
	if _, ok := evt.(*events.Connected); ok {
		if client.Store.ID != nil {
			senderID := client.Store.ID.User
			// Clear reconnecting flag
			cm.mu.Lock()
			delete(cm.reconnecting, senderID)
			cm.mu.Unlock()

			// Mark sender as active when reconnected
			if err := repository.UpdateSenderStatus(cm.db, senderID, true); err != nil {
				log.Printf("Failed to update sender status to active for %s: %v", senderID, err)
			} else {
				log.Printf("✓ Client %s connected and marked as active", senderID)
			}
		}
	}

	// Handle disconnected events - let whatsmeow handle automatic reconnection
	// Do NOT manually reconnect - whatsmeow has built-in reconnection logic
	// Manual reconnection can trigger WhatsApp's security system
	if _, ok := evt.(*events.Disconnected); ok {
		if client.Store.ID != nil {
			senderID := client.Store.ID.User
			log.Printf("Client %s disconnected - whatsmeow will handle automatic reconnection", senderID)
			// Don't manually reconnect - whatsmeow handles this internally
		}
	}

	// Handle stream error events - these usually recover automatically via whatsmeow
	if streamErr, ok := evt.(*events.StreamError); ok {
		if client.Store.ID != nil {
			senderID := client.Store.ID.User
			log.Printf("⚠ Client %s stream error (code: %s) - whatsmeow will handle recovery", senderID, streamErr.Code)
			// Don't manually intervene - let whatsmeow handle it
		}
	}

	// Handle logout events with cleanup - ONLY for explicit logouts
	if logoutEvt, ok := evt.(*events.LoggedOut); ok {
		if client.Store.ID != nil {
			senderID := client.Store.ID.User

			// Reason is a ConnectFailureReason enum
			reason := logoutEvt.Reason
			log.Printf("[ClientManager] Client %s logged out - Reason: %d (%s)", senderID, reason, reason.String())

			// For ANY logout event, clean up properly
			// WhatsApp logged out this device - we should NOT try to reconnect
			// Reconnection attempts can trigger WhatsApp's security system
			log.Printf("Device %s was logged out by WhatsApp, cleaning up session", senderID)

			// Update sender status to inactive
			if err := repository.UpdateSenderStatus(cm.db, senderID, false); err != nil {
				log.Printf("Failed to update sender status for %s: %v", senderID, err)
			}

			// Remove from clients map
			cm.mu.Lock()
			delete(cm.clients, senderID)
			delete(cm.reconnecting, senderID)

			// If this was the default sender, clear it
			if cm.defaultSenderID == senderID {
				cm.defaultSenderID = ""
				log.Printf("Default sender %s was logged out, clearing default", senderID)
			}
			cm.mu.Unlock()

			log.Printf("Client %s removed from active clients", senderID)

			// Delete the device session from database - session is invalid now
			if err := cm.container.DeleteDevice(context.Background(), client.Store); err != nil {
				log.Printf("Failed to delete device session for %s: %v", senderID, err)
			} else {
				log.Printf("Device session deleted for %s", senderID)
			}

			log.Printf("⚠ To reconnect sender %s, please re-register the device via QR code or pairing code", senderID)
		}
	}

	// Handle stream replaced events - another session took over
	// Do NOT try to reconnect - this will cause a reconnection loop
	if _, ok := evt.(*events.StreamReplaced); ok {
		if client.Store.ID != nil {
			senderID := client.Store.ID.User
			log.Printf("⚠ Client %s - stream replaced by another session (do not reconnect)", senderID)
			// Don't reconnect - another session has taken over
		}
	}

	// Call the regular event handler for all events
	handleEvent(evt, cm.db, client)
}

// Note: Removed attemptReconnect function - whatsmeow handles reconnection internally
// Manual reconnection attempts can trigger WhatsApp's security system and cause
// devices to be logged out with "unexpected issue" errors// RemoveClient removes a client and marks it as inactive
func (cm *ClientManager) RemoveClient(senderID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	client, exists := cm.clients[senderID]
	if !exists {
		return fmt.Errorf("client not found: %s", senderID)
	}

	// Disconnect the client
	client.Disconnect()

	// Update sender status to inactive
	if err := repository.UpdateSenderStatus(cm.db, senderID, false); err != nil {
		log.Printf("Failed to update sender status for %s: %v", senderID, err)
	}

	// Delete from clients map
	delete(cm.clients, senderID)

	// If this was the default sender, clear it
	if cm.defaultSenderID == senderID {
		cm.defaultSenderID = ""
	}

	// Delete the device session
	if err := cm.container.DeleteDevice(context.Background(), client.Store); err != nil {
		return fmt.Errorf("failed to delete device session: %w", err)
	}

	log.Printf("Client %s removed successfully", senderID)
	return nil
}

// GetContainer returns the sqlstore container for creating new devices
func (cm *ClientManager) GetContainer() *sqlstore.Container {
	return cm.container
}

// SetDefaultSender sets the default sender in both memory and database
func (cm *ClientManager) SetDefaultSender(senderID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if client exists
	if _, exists := cm.clients[senderID]; !exists {
		return fmt.Errorf("sender not found: %s", senderID)
	}

	// Update database
	if err := repository.SetDefaultSender(cm.db, senderID); err != nil {
		return fmt.Errorf("failed to set default sender in database: %w", err)
	}

	// Update in-memory default
	cm.defaultSenderID = senderID
	log.Printf("Default sender set to: %s", senderID)

	return nil
}

// GetDefaultSenderID returns the current default sender ID
func (cm *ClientManager) GetDefaultSenderID() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.defaultSenderID
}

// AddNewClient registers a new WhatsApp client for a new phone number
// IMPORTANT: Each call creates a NEW sender with a DIFFERENT WhatsApp phone number.
// WhatsApp limits each phone number to 4 linked devices (including the phone itself).
// To have multiple senders, you need multiple WhatsApp accounts (different phone numbers).
// Example: Sender1 (+1234567890), Sender2 (+9876543210), Sender3 (+5555555555)
func (cm *ClientManager) AddNewClient() (*whatsmeow.Client, error) {
	// Create a NEW device store for the new phone number
	// NOTE: Do NOT use GetFirstDevice() - that returns existing devices
	deviceStore := cm.container.NewDevice()

	logLevel := GetLogLevel()
	clientLog := waLog.Stdout("NewClient", logLevel, true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Create channels to wait for pairing success and connection
	pairingDone := make(chan bool, 1)
	connectionDone := make(chan bool, 1)
	pairingFailed := make(chan bool, 1)
	pairingTimeout := time.After(5 * time.Minute)

	// Add event handler to track connection status
	eventID := client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.PairSuccess:
			fmt.Println("\n✓ QR code scanned successfully! Waiting for connection to complete...")
			pairingDone <- true
		case *events.Connected:
			fmt.Println("✓ Connection established!")
			connectionDone <- true
		case *events.LoggedOut:
			fmt.Println("\n✗ Login failed or logged out")
			pairingFailed <- true
		default:
			// Also handle regular events
			handleEvent(v, cm.db, client)
		}
	})
	defer client.RemoveEventHandler(eventID)

	// Check if this device is already registered (shouldn't be for new device)
	if client.Store.ID != nil {
		return nil, fmt.Errorf("device already has an ID - this shouldn't happen for a new device")
	}

	// Get QR code for scanning
	fmt.Println("\n=== Adding New WhatsApp Phone Number ===")
	fmt.Println("Please scan this QR code with the WhatsApp account you want to add:")
	fmt.Println()

	qrChan, _ := client.GetQRChannel(context.Background())
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Display QR codes as they come
	go func() {
		for evt := range qrChan {
			if evt.Event == "code" {
				// Display QR code in terminal
				fmt.Println("QR Code (scan with WhatsApp):")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println()
			} else {
				fmt.Printf("Login event: %s\n", evt.Event)
			}
		}
	}()

	// Wait for pairing to complete
	fmt.Println("Waiting for QR code scan...")
	select {
	case <-pairingDone:
		// Pairing successful, now wait for connection
		fmt.Println("Waiting for WhatsApp connection to complete...")
		select {
		case <-connectionDone:
			// Connection successful
			fmt.Println("✓ Successfully connected new phone number!")
		case <-time.After(30 * time.Second):
			return nil, fmt.Errorf("timeout waiting for connection after pairing")
		case <-pairingFailed:
			return nil, fmt.Errorf("connection failed after pairing")
		}
	case <-pairingFailed:
		return nil, fmt.Errorf("pairing failed")
	case <-pairingTimeout:
		return nil, fmt.Errorf("timeout waiting for QR code scan (5 minutes)")
	}

	// Wait for device ID to be set
	if client.Store.ID == nil {
		return nil, fmt.Errorf("failed to get device ID after connection")
	}

	senderID := client.Store.ID.User
	fmt.Printf("✓ New sender registered with ID: %s\n", senderID)

	// Register sender in database
	cm.ensureSenderRecord(senderID, client.Store.ID.User)

	// Add to client map
	cm.mu.Lock()
	cm.clients[senderID] = client
	cm.mu.Unlock()

	fmt.Println("✓ New phone number is ready to send messages!")

	return client, nil
}

// AddNewClientWithPairingCode registers a new WhatsApp client using phone number pairing code
// This method sends a pairing code via SMS instead of using QR scanning
// IMPORTANT: Each call creates a NEW sender with a DIFFERENT WhatsApp phone number.
// WhatsApp limits each phone number to 4 linked devices (including the phone itself).
// To have multiple senders, you need multiple WhatsApp accounts (different phone numbers).
// Example: Sender1 (+1234567890), Sender2 (+9876543210), Sender3 (+5555555555)
func (cm *ClientManager) AddNewClientWithPairingCode(phoneNumber string) (*whatsmeow.Client, error) {
	// Create a NEW device store for the new phone number
	deviceStore := cm.container.NewDevice()

	logLevel := GetLogLevel()
	clientLog := waLog.Stdout("NewClient", logLevel, true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Add event handler with client manager awareness
	client.AddEventHandler(func(evt interface{}) {
		cm.handleEventWithCleanup(evt, client)
	})

	// Check if this device is already registered
	if client.Store.ID != nil {
		return nil, fmt.Errorf("device already has an ID - this shouldn't happen for a new device")
	}

	fmt.Printf("\n=== Adding WhatsApp Phone Number: %s ===\n", phoneNumber)
	fmt.Println("Connecting to WhatsApp...")
	fmt.Println()

	// Connect first (required before requesting pairing code)
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	fmt.Println("✓ Connected! Requesting pairing code via SMS...")
	fmt.Println()

	// Request pairing code (will be sent via SMS to the phone number)
	code, err := client.PairPhone(context.Background(), phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		return nil, fmt.Errorf("failed to request pairing code: %w", err)
	}

	fmt.Printf("✓ Pairing code sent to %s: %s\n", phoneNumber, code)
	fmt.Println()
	fmt.Println("Enter this code in your WhatsApp app:")
	fmt.Println("  1. Open WhatsApp on your phone")
	fmt.Println("  2. Go to Settings > Linked Devices")
	fmt.Println("  3. Tap 'Link a Device'")
	fmt.Println("  4. Tap 'Link with phone number instead'")
	fmt.Printf("  5. Enter the code: %s\n", code)
	fmt.Println()
	fmt.Println("Waiting for pairing to complete...")

	// Create channels to wait for pairing success and connection
	pairingDone := make(chan bool, 1)
	connectionDone := make(chan bool, 1)
	pairingFailed := make(chan bool, 1)
	pairingTimeout := time.After(5 * time.Minute) // 5 minute timeout

	// Add event handler to detect successful pairing and connection
	eventID := client.AddEventHandler(func(evt interface{}) {
		switch evt.(type) {
		case *events.PairSuccess:
			fmt.Println("\n✓ Pairing successful! Waiting for connection to complete...")
			pairingDone <- true
		case *events.Connected:
			fmt.Println("✓ Connection established!")
			connectionDone <- true
		case *events.LoggedOut:
			fmt.Println("\n✗ Pairing failed - logged out")
			pairingFailed <- true
		}
	})
	defer client.RemoveEventHandler(eventID)

	// Wait for pairing completion or timeout
	select {
	case <-pairingDone:
		// Pairing successful, now wait for connection
		fmt.Println("Waiting for WhatsApp connection to complete...")
		select {
		case <-connectionDone:
			// Connection successful
			fmt.Println("✓ Successfully connected!")
		case <-time.After(30 * time.Second):
			return nil, fmt.Errorf("timeout waiting for connection after pairing")
		case <-pairingFailed:
			return nil, fmt.Errorf("connection failed after pairing")
		}
	case <-pairingFailed:
		return nil, fmt.Errorf("pairing failed")
	case <-pairingTimeout:
		return nil, fmt.Errorf("pairing timed out after 5 minutes")
	}

	// Wait for device ID to be set (indicates successful pairing)
	if client.Store.ID == nil {
		return nil, fmt.Errorf("pairing not completed - device ID not set")
	}

	senderID := client.Store.ID.User
	fmt.Printf("\n✓ Successfully paired! Sender ID: %s\n", senderID)

	// Register sender in database
	cm.ensureSenderRecord(senderID, phoneNumber)

	// Add to client map
	cm.mu.Lock()
	cm.clients[senderID] = client
	cm.mu.Unlock()

	fmt.Println("✓ New phone number is ready to send messages!")

	return client, nil
}
