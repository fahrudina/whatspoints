package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/wa-serv/internal/domain"
)

// MockWhatsAppRepository is a mock implementation of WhatsAppRepository
type MockWhatsAppRepository struct {
	mock.Mock
}

func (m *MockWhatsAppRepository) SendMessage(ctx context.Context, to, message string) (*domain.Message, error) {
	args := m.Called(ctx, to, message)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockWhatsAppRepository) SendMessageFrom(ctx context.Context, from, to, message string) (*domain.Message, error) {
	args := m.Called(ctx, from, to, message)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *MockWhatsAppRepository) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWhatsAppRepository) IsLoggedIn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWhatsAppRepository) GetJID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockWhatsAppRepository) GetSenderJID(senderID string) (string, error) {
	args := m.Called(senderID)
	return args.String(0), args.Error(1)
}

func (m *MockWhatsAppRepository) ListSenders() ([]*domain.Sender, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Sender), args.Error(1)
}

func (m *MockWhatsAppRepository) GetDefaultSender() (*domain.Sender, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Sender), args.Error(1)
}

// MockMessageService is a mock implementation of MessageService
type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) SendMessage(ctx context.Context, req *domain.SendMessageRequest) (*domain.SendMessageResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SendMessageResponse), args.Error(1)
}

func (m *MockMessageService) GetStatus(ctx context.Context) (*domain.ServiceStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ServiceStatus), args.Error(1)
}

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateCredentials(username, password string) bool {
	args := m.Called(username, password)
	return args.Bool(0)
}
