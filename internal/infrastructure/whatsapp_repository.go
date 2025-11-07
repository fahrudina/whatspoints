package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/wa-serv/internal/domain"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type whatsappRepository struct {
	client    *whatsmeow.Client // Default client for backward compatibility
	db        *sql.DB
	clientMap map[string]*whatsmeow.Client // Map of sender_id -> client
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

// RegisterClient registers a client for a specific sender
func (r *whatsappRepository) RegisterClient(senderID string, client *whatsmeow.Client) {
	r.clientMap[senderID] = client
}

// SendMessage sends a WhatsApp message using the default client
func (r *whatsappRepository) SendMessage(ctx context.Context, to, message string) (*domain.Message, error) {
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
	resp, err := r.client.SendMessage(ctx, jid, msg)
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
	// Get the client for this sender
	client, ok := r.clientMap[from]
	if !ok {
		return nil, fmt.Errorf("sender not found: %s", from)
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
	return r.client.IsConnected()
}

// IsLoggedIn checks if WhatsApp client is logged in
func (r *whatsappRepository) IsLoggedIn() bool {
	return r.client.IsLoggedIn()
}

// GetJID gets the WhatsApp JID
func (r *whatsappRepository) GetJID() string {
	if r.client.Store.ID != nil {
		return r.client.Store.ID.String()
	}
	return ""
}

// GetSenderJID gets the WhatsApp JID for a specific sender
func (r *whatsappRepository) GetSenderJID(senderID string) (string, error) {
	client, ok := r.clientMap[senderID]
	if !ok {
		return "", domain.ErrSenderNotFound
	}
	if client.Store.ID != nil {
		return client.Store.ID.String(), nil
	}
	return "", nil
}

// ListSenders returns all available senders
func (r *whatsappRepository) ListSenders() ([]*domain.Sender, error) {
	if r.db == nil {
		// Return default sender only if no database
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

	query := `SELECT sender_id, phone_number, name, is_default, is_active FROM senders WHERE is_active = TRUE`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query senders: %w", err)
	}
	defer rows.Close()

	var senders []*domain.Sender
	for rows.Next() {
		var s domain.Sender
		if err := rows.Scan(&s.ID, &s.PhoneNumber, &s.Name, &s.IsDefault, &s.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan sender: %w", err)
		}
		senders = append(senders, &s)
	}

	return senders, nil
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

	query := `SELECT sender_id, phone_number, name, is_default, is_active FROM senders WHERE is_default = TRUE AND is_active = TRUE LIMIT 1`
	var s domain.Sender
	err := r.db.QueryRow(query).Scan(&s.ID, &s.PhoneNumber, &s.Name, &s.IsDefault, &s.IsActive)
	if err == sql.ErrNoRows {
		// No default sender set, return error indicating no default is available
		return nil, domain.ErrNoActiveSender
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get default sender: %w", err)
	}

	return &s, nil
}
