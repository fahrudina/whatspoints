package infrastructure

import (
	"context"
	"fmt"

	"github.com/wa-serv/internal/domain"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type whatsappRepository struct {
	client *whatsmeow.Client
}

// NewWhatsAppRepository creates a new WhatsApp repository
func NewWhatsAppRepository(client *whatsmeow.Client) domain.WhatsAppRepository {
	return &whatsappRepository{
		client: client,
	}
}

// SendMessage sends a WhatsApp message
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
