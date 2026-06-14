package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wa-serv/config"
	"github.com/wa-serv/internal/domain"
	"github.com/wa-serv/internal/mocks"
)

func TestAIService_Disabled_ReturnsErrorWithoutCallingClient(t *testing.T) {
	client := &mocks.MockAIClient{}
	svc := NewAIService(client, config.AIConfig{Enabled: false})

	resp, err := svc.GenerateReply(context.Background(), &domain.AIReplyRequest{Message: "hi"})

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrAIResponseDisabled)
	client.AssertNotCalled(t, "GenerateReply", mock.Anything, mock.Anything, mock.Anything)
}

func TestAIService_EmptyMessage_Rejected(t *testing.T) {
	client := &mocks.MockAIClient{}
	svc := NewAIService(client, config.AIConfig{Enabled: true})

	resp, err := svc.GenerateReply(context.Background(), &domain.AIReplyRequest{Message: "   "})

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrEmptyMessage)
	client.AssertNotCalled(t, "GenerateReply", mock.Anything, mock.Anything, mock.Anything)
}

func TestAIService_Enabled_CallsClient(t *testing.T) {
	client := &mocks.MockAIClient{}
	expected := &domain.AIReplyResponse{Reply: "Masih kak", Intent: "ask_promo"}
	client.On("GenerateReply", mock.Anything, "promo?", "628").Return(expected, nil)

	svc := NewAIService(client, config.AIConfig{Enabled: true})
	resp, err := svc.GenerateReply(context.Background(), &domain.AIReplyRequest{Message: "promo?", PhoneNumber: "628"})

	assert.NoError(t, err)
	assert.Equal(t, expected, resp)
	client.AssertExpectations(t)
}
