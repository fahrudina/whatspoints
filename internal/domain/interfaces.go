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
)

// WhatsAppRepository defines the interface for WhatsApp operations
type WhatsAppRepository interface {
	SendMessage(ctx context.Context, to, message string) (*Message, error)
	IsConnected() bool
	IsLoggedIn() bool
	GetJID() string
}

// MessageService defines the business logic interface for messaging
type MessageService interface {
	SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error)
	GetStatus(ctx context.Context) (*ServiceStatus, error)
}

// AuthService defines the authentication interface
type AuthService interface {
	ValidateCredentials(username, password string) bool
}
