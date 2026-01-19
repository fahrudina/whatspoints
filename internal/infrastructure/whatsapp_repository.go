package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/wa-serv/internal/domain"
	"github.com/wa-serv/repository"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type whatsappRepository struct {
	client        *whatsmeow.Client // Default client for backward compatibility
	db            *sql.DB
	clientMap     map[string]*whatsmeow.Client // Map of sender_id -> client
	mu            sync.RWMutex                 // Protects clientMap
	clientManager interface {                  // Interface to get clients dynamically
		GetClient(senderID string) (*whatsmeow.Client, error)
		GetDefaultClient() (*whatsmeow.Client, error)
		GetAllClients() map[string]*whatsmeow.Client
	}
}

// NewWhatsAppRepository creates a new WhatsApp repository
func NewWhatsAppRepository(client *whatsmeow.Client) domain.WhatsAppRepository {
	return &whatsappRepository{
		client:    client,
		clientMap: make(map[string]*whatsmeow.Client),
	}
}

// NewWhatsAppRepositoryWithDB creates a new WhatsApp repository with database support
func NewWhatsAppRepositoryWithDB(client *whatsmeow.Client, db *sql.DB) domain.WhatsAppRepository {
	return &whatsappRepository{
		client:    client,
		db:        db,
		clientMap: make(map[string]*whatsmeow.Client),
	}
}

// NewWhatsAppRepositoryWithClients creates a new WhatsApp repository with multiple clients
func NewWhatsAppRepositoryWithClients(defaultClient *whatsmeow.Client, db *sql.DB, clients map[string]*whatsmeow.Client) domain.WhatsAppRepository {
	repo := &whatsappRepository{
		client:    defaultClient,
		db:        db,
		clientMap: make(map[string]*whatsmeow.Client),
	}

	// Register all clients
	for senderID, client := range clients {
		repo.clientMap[senderID] = client
	}

	return repo
}

// NewWhatsAppRepositoryWithClientManager creates a repository that uses ClientManager dynamically
func NewWhatsAppRepositoryWithClientManager(db *sql.DB, clientManager interface {
	GetClient(senderID string) (*whatsmeow.Client, error)
	GetDefaultClient() (*whatsmeow.Client, error)
	GetAllClients() map[string]*whatsmeow.Client
}) domain.WhatsAppRepository {
	// Try to get default client, but don't fail if it's not available yet
	// The repository will handle nil client gracefully via getClient accessor
	defaultClient, err := clientManager.GetDefaultClient()
	if err != nil {
		// Log but continue - getClient will handle getting a valid client later
		fmt.Printf("No default client available during repository initialization: %v\n", err)
		defaultClient = nil
	}
	return &whatsappRepository{
		client:        defaultClient,
		db:            db,
		clientMap:     make(map[string]*whatsmeow.Client),
		clientManager: clientManager,
	}
}

// RegisterClient registers a client for a specific sender
func (r *whatsappRepository) RegisterClient(senderID string, client *whatsmeow.Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clientMap[senderID] = client
}

// getClient safely retrieves a client, falling back to defaults if necessary
func (r *whatsappRepository) getClient(senderID string) (*whatsmeow.Client, error) {
	// If a specific sender is requested, try to get that client
	if senderID != "" {
		if r.clientManager != nil {
			client, err := r.clientManager.GetClient(senderID)
			if err == nil && client != nil {
				return client, nil
			}
		}
		r.mu.RLock()
		client, ok := r.clientMap[senderID]
		r.mu.RUnlock()
		if ok && client != nil {
			return client, nil
		}
		// Specific sender was requested but not found - return error instead of falling back
		return nil, domain.ErrSenderNotFound
	}

	// Fall back to default client from repository (only when senderID == "")
	if r.client != nil {
		return r.client, nil
	}

	// Try to get default client from manager
	if r.clientManager != nil {
		client, err := r.clientManager.GetDefaultClient()
		if err == nil && client != nil {
			return client, nil
		}
	}

	return nil, fmt.Errorf("no WhatsApp client available")
}

// SendMessage sends a WhatsApp message using the default client
func (r *whatsappRepository) SendMessage(ctx context.Context, to, message string) (*domain.Message, error) {
	// Get a valid client
	client, err := r.getClient("")
	if err != nil {
		return nil, fmt.Errorf("no client available: %w", err)
	}

	// Parse JID
	jid, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JID: %w", err)
	}

	// Create WhatsApp message
	msg := &waProto.Message{
		Conversation: proto.String(message),
	}

	// Send message
	resp, err := client.SendMessage(ctx, jid, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &domain.Message{
		ID:      resp.ID,
		To:      to,
		Content: message,
		SentAt:  resp.Timestamp.String(),
	}, nil
}

// SendMessageFrom sends a WhatsApp message from a specific sender
func (r *whatsappRepository) SendMessageFrom(ctx context.Context, from, to, message string) (*domain.Message, error) {
	// Use getClient helper to safely retrieve the client with proper nil checks
	client, err := r.getClient(from)
	if err != nil {
		return nil, fmt.Errorf("sender not found or not initialized: %s", from)
	}

	// Check if client is connected
	if !client.IsConnected() {
		return nil, fmt.Errorf("sender %s is not connected", from)
	}

	// Parse JID
	jid, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JID: %w", err)
	}

	// Create WhatsApp message
	msg := &waProto.Message{
		Conversation: proto.String(message),
	}

	// Send message
	resp, err := client.SendMessage(ctx, jid, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &domain.Message{
		ID:      resp.ID,
		To:      to,
		Content: message,
		SentAt:  resp.Timestamp.String(),
	}, nil
}

// IsConnected checks if WhatsApp client is connected
func (r *whatsappRepository) IsConnected() bool {
	// If we have a client manager, check if any client is connected
	if r.clientManager != nil {
		clients := r.clientManager.GetAllClients()
		for _, client := range clients {
			// Guard against nil clients
			if client != nil && client.IsConnected() {
				return true
			}
		}
		return false
	}

	// Fallback to default client check
	if r.client != nil {
		return r.client.IsConnected()
	}

	return false
}

// IsLoggedIn checks if WhatsApp client is logged in
func (r *whatsappRepository) IsLoggedIn() bool {
	client, err := r.getClient("")
	if err != nil || client == nil {
		return false
	}
	return client.IsLoggedIn()
}

// GetJID gets the WhatsApp JID
func (r *whatsappRepository) GetJID() string {
	client, err := r.getClient("")
	if err != nil || client == nil {
		return ""
	}
	if client.Store.ID != nil {
		return client.Store.ID.String()
	}
	return ""
}

// GetSenderJID gets the WhatsApp JID for a specific sender
func (r *whatsappRepository) GetSenderJID(senderID string) (string, error) {
	client, err := r.getClient(senderID)
	if err != nil {
		return "", domain.ErrSenderNotFound
	}
	if client.Store.ID != nil {
		return client.Store.ID.String(), nil
	}
	return "", nil
}

// ListSenders returns all active senders
func (r *whatsappRepository) ListSenders() ([]*domain.Sender, error) {
	if r.db == nil {
		// Return default sender if no database
		return []*domain.Sender{
			{
				ID:          "default",
				PhoneNumber: r.GetJID(),
				Name:        "Default Sender",
				IsDefault:   true,
				IsActive:    r.IsConnected(),
			},
		}, nil
	}

	// Use repository layer
	senders, err := repository.GetAllSenders(r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get senders: %w", err)
	}

	// Convert repository.Sender to domain.Sender
	domainSenders := make([]*domain.Sender, 0, len(senders))
	for _, s := range senders {
		if s.IsActive {
			domainSenders = append(domainSenders, &domain.Sender{
				ID:          s.SenderID,
				PhoneNumber: s.PhoneNumber,
				Name:        s.Name,
				IsDefault:   s.IsDefault,
				IsActive:    s.IsActive,
			})
		}
	}

	return domainSenders, nil
}

// GetDefaultSender returns the default sender
func (r *whatsappRepository) GetDefaultSender() (*domain.Sender, error) {
	if r.db == nil {
		// Return default sender if no database
		return &domain.Sender{
			ID:          "default",
			PhoneNumber: r.GetJID(),
			Name:        "Default Sender",
			IsDefault:   true,
			IsActive:    r.IsConnected(),
		}, nil
	}

	// Get all senders and find the default one
	senders, err := repository.GetAllSenders(r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get senders: %w", err)
	}

	// Find default and active sender
	for _, s := range senders {
		if s.IsDefault && s.IsActive {
			return &domain.Sender{
				ID:          s.SenderID,
				PhoneNumber: s.PhoneNumber,
				Name:        s.Name,
				IsDefault:   s.IsDefault,
				IsActive:    s.IsActive,
			}, nil
		}
	}

	// No default sender set
	return nil, domain.ErrNoActiveSender
}
