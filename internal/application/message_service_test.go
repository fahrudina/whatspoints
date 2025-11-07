package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wa-serv/internal/domain"
	"github.com/wa-serv/internal/mocks"
)

func TestMessageService_SendMessage_Success(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockWhatsAppRepository{}
	service := NewMessageService(mockRepo)

	req := &domain.SendMessageRequest{
		To:      "+1234567890",
		Message: "Test message",
	}

	expectedMessage := &domain.Message{
		ID:      "test-id",
		To:      "1234567890@s.whatsapp.net",
		Content: "Test message",
		SentAt:  "2023-01-01",
	}

	mockRepo.On("IsConnected").Return(true)
	mockRepo.On("SendMessage", mock.Anything, "1234567890@s.whatsapp.net", "Test message").Return(expectedMessage, nil)

	// Act
	response, err := service.SendMessage(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Message sent successfully", response.Message)
	assert.Equal(t, "test-id", response.ID)

	mockRepo.AssertExpectations(t)
}

func TestMessageService_SendMessage_NotConnected(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockWhatsAppRepository{}
	service := NewMessageService(mockRepo)

	req := &domain.SendMessageRequest{
		To:      "+1234567890",
		Message: "Test message",
	}

	mockRepo.On("IsConnected").Return(false)

	// Act
	response, err := service.SendMessage(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrWhatsAppNotConnected, err)
	assert.False(t, response.Success)
	assert.Equal(t, "WhatsApp client is not connected", response.Message)

	mockRepo.AssertExpectations(t)
}

func TestMessageService_SendMessage_InvalidPhoneNumber(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockWhatsAppRepository{}
	service := NewMessageService(mockRepo)

	req := &domain.SendMessageRequest{
		To:      "123", // Too short
		Message: "Test message",
	}

	mockRepo.On("IsConnected").Return(true)

	// Act
	response, err := service.SendMessage(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidPhoneNumber, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Invalid phone number format", response.Message)

	mockRepo.AssertExpectations(t)
}

func TestMessageService_SendMessage_EmptyRequest(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockWhatsAppRepository{}
	service := NewMessageService(mockRepo)

	req := &domain.SendMessageRequest{
		To:      "",
		Message: "",
	}

	// Act
	response, err := service.SendMessage(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "phone number is required")
}

func TestMessageService_GetStatus_Success(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockWhatsAppRepository{}
	service := NewMessageService(mockRepo)

	mockRepo.On("IsConnected").Return(true)
	mockRepo.On("IsLoggedIn").Return(true)
	mockRepo.On("GetJID").Return("test@s.whatsapp.net")

	// Act
	status, err := service.GetStatus(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, status.WhatsApp.Connected)
	assert.True(t, status.WhatsApp.LoggedIn)
	assert.Equal(t, "test@s.whatsapp.net", status.WhatsApp.JID)

	mockRepo.AssertExpectations(t)
}

func TestMessageService_FormatPhoneNumber(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockWhatsAppRepository{}
	service := &messageService{whatsappRepo: mockRepo}

	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Valid phone with plus",
			input:    "+1234567890",
			expected: "1234567890@s.whatsapp.net",
			hasError: false,
		},
		{
			name:     "Valid phone without plus",
			input:    "1234567890",
			expected: "1234567890@s.whatsapp.net",
			hasError: false,
		},
		{
			name:     "Phone with spaces and dashes",
			input:    "+1 (234) 567-890",
			expected: "1234567890@s.whatsapp.net",
			hasError: false,
		},
		{
			name:     "Too short phone",
			input:    "123",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result, err := service.formatPhoneNumber(tt.input)

			// Assert
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMessageService_SendMessage_WithSender_Success(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockWhatsAppRepository{}
	service := NewMessageService(mockRepo)

	req := &domain.SendMessageRequest{
		To:      "+1234567890",
		Message: "Test message",
		From:    "sender123",
	}

	expectedMessage := &domain.Message{
		ID:      "test-id",
		To:      "1234567890@s.whatsapp.net",
		Content: "Test message",
		SentAt:  "2023-01-01",
	}

	mockRepo.On("IsConnected").Return(true)
	mockRepo.On("SendMessageFrom", mock.Anything, "sender123", "1234567890@s.whatsapp.net", "Test message").Return(expectedMessage, nil)

	// Act
	response, err := service.SendMessage(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Message sent successfully", response.Message)
	assert.Equal(t, "test-id", response.ID)

	mockRepo.AssertExpectations(t)
}

func TestMessageService_SendMessage_WithSender_NotFound(t *testing.T) {
	// Arrange
	mockRepo := &mocks.MockWhatsAppRepository{}
	service := NewMessageService(mockRepo)

	req := &domain.SendMessageRequest{
		To:      "+1234567890",
		Message: "Test message",
		From:    "nonexistent",
	}

	mockRepo.On("IsConnected").Return(true)
	mockRepo.On("SendMessageFrom", mock.Anything, "nonexistent", "1234567890@s.whatsapp.net", "Test message").Return(nil, domain.ErrSenderNotFound)

	// Act
	response, err := service.SendMessage(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrMessageSendFailed, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Failed to send message")

	mockRepo.AssertExpectations(t)
}
