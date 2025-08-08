package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wa-serv/internal/domain"
)

type messageService struct {
	whatsappRepo domain.WhatsAppRepository
}

// NewMessageService creates a new message service
func NewMessageService(whatsappRepo domain.WhatsAppRepository) domain.MessageService {
	return &messageService{
		whatsappRepo: whatsappRepo,
	}
}

// SendMessage implements the business logic for sending messages
func (s *messageService) SendMessage(ctx context.Context, req *domain.SendMessageRequest) (*domain.SendMessageResponse, error) {
	// Validate input
	if err := s.validateSendMessageRequest(req); err != nil {
		return &domain.SendMessageResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// Check if WhatsApp is connected
	if !s.whatsappRepo.IsConnected() {
		return &domain.SendMessageResponse{
			Success: false,
			Message: "WhatsApp client is not connected",
		}, domain.ErrWhatsAppNotConnected
	}

	// Format phone number
	formattedPhone, err := s.formatPhoneNumber(req.To)
	if err != nil {
		return &domain.SendMessageResponse{
			Success: false,
			Message: "Invalid phone number format",
		}, domain.ErrInvalidPhoneNumber
	}

	// Create a context with timeout to prevent hanging
	sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Send message
	message, err := s.whatsappRepo.SendMessage(sendCtx, formattedPhone, req.Message)
	if err != nil {
		return &domain.SendMessageResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to send message: %v", err),
		}, domain.ErrMessageSendFailed
	}

	return &domain.SendMessageResponse{
		Success: true,
		Message: "Message sent successfully",
		ID:      message.ID,
	}, nil
}

// GetStatus implements the business logic for getting service status
func (s *messageService) GetStatus(ctx context.Context) (*domain.ServiceStatus, error) {
	whatsappStatus := domain.WhatsAppStatus{
		Connected: s.whatsappRepo.IsConnected(),
		LoggedIn:  s.whatsappRepo.IsLoggedIn(),
		JID:       s.whatsappRepo.GetJID(),
	}

	return &domain.ServiceStatus{
		WhatsApp: whatsappStatus,
	}, nil
}

// validateSendMessageRequest validates the send message request
func (s *messageService) validateSendMessageRequest(req *domain.SendMessageRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if strings.TrimSpace(req.To) == "" {
		return fmt.Errorf("recipient phone number is required")
	}

	if strings.TrimSpace(req.Message) == "" {
		return fmt.Errorf("message content is required")
	}

	return nil
}

// formatPhoneNumber formats and validates phone number
func (s *messageService) formatPhoneNumber(phone string) (string, error) {
	phone = strings.TrimSpace(phone)

	// Remove any spaces, dashes, or other non-numeric characters except +
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	// Remove + if present since WhatsApp JIDs don't use +
	phone = strings.TrimPrefix(phone, "+")

	// Basic validation - should be at least 10 digits
	if len(phone) < 10 {
		return "", fmt.Errorf("phone number too short")
	}

	// Ensure it's all digits
	for _, char := range phone {
		if char < '0' || char > '9' {
			return "", fmt.Errorf("phone number contains invalid characters")
		}
	}

	// Add WhatsApp suffix if not present
	if !strings.HasSuffix(phone, "@s.whatsapp.net") {
		phone = phone + "@s.whatsapp.net"
	}

	return phone, nil
}
