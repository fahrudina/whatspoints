package domain

import (
	"context"
	"errors"
)

// Common errors
var (
	ErrWhatsAppNotConnected = errors.New("whatsapp client is not connected")
	ErrInvalidPhoneNumber   = errors.New("invalid phone number format")
	ErrMessageSendFailed    = errors.New("failed to send message")
	ErrUnauthorized         = errors.New("unauthorized access")
	ErrSenderNotFound       = errors.New("sender not found")
	ErrNoActiveSender       = errors.New("no active sender available")
)

// WhatsAppRepository defines the interface for WhatsApp operations
type WhatsAppRepository interface {
	SendMessage(ctx context.Context, to, message string) (*Message, error)
	SendMessageFrom(ctx context.Context, from, to, message string) (*Message, error)
	IsConnected() bool
	IsLoggedIn() bool
	GetJID() string
	GetSenderJID(senderID string) (string, error)
	ListSenders() ([]*Sender, error)
	GetDefaultSender() (*Sender, error)
}

// MessageService defines the business logic interface for messaging
type MessageService interface {
	SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error)
	GetStatus(ctx context.Context) (*ServiceStatus, error)
	ListSenders(ctx context.Context) ([]*Sender, error)
}

// AuthService defines the authentication interface
type AuthService interface {
	ValidateCredentials(username, password string) bool
}
