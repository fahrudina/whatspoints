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
	ErrAIResponseDisabled   = errors.New("AI response feature is disabled")
	ErrEmptyMessage         = errors.New("message is required")
)

// AIClient talks to the external AI sidecar service over HTTP.
type AIClient interface {
	GenerateReply(ctx context.Context, message, phoneNumber string) (*AIReplyResponse, error)
}

// AIService is the business logic for generating suggested AI replies.
type AIService interface {
	GenerateReply(ctx context.Context, req *AIReplyRequest) (*AIReplyResponse, error)
}

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

// SenderRegistrationService defines the business logic interface for sender registration
type SenderRegistrationService interface {
	StartQRRegistration(ctx context.Context) (*RegisterSenderQRResponse, error)
	StartCodeRegistration(ctx context.Context, req *RegisterSenderCodeRequest) (*RegisterSenderCodeResponse, error)
	GetRegistrationStatus(ctx context.Context, sessionID string) (*RegistrationStatusResponse, error)
}

// AuthService defines the authentication interface
type AuthService interface {
	ValidateCredentials(username, password string) bool
}
