package application

import (
	"context"
	"strings"

	"github.com/wa-serv/config"
	"github.com/wa-serv/internal/domain"
)

type aiService struct {
	client domain.AIClient
	cfg    config.AIConfig
}

// NewAIService creates the AI reply-suggestion service. When the feature is
// disabled the client may be nil; GenerateReply short-circuits before using it.
func NewAIService(client domain.AIClient, cfg config.AIConfig) domain.AIService {
	return &aiService{client: client, cfg: cfg}
}

// GenerateReply validates input and asks the AI sidecar for a suggested reply.
// It never sends a WhatsApp message — this phase only produces suggestions.
func (s *aiService) GenerateReply(ctx context.Context, req *domain.AIReplyRequest) (*domain.AIReplyResponse, error) {
	if !s.cfg.Enabled {
		return nil, domain.ErrAIResponseDisabled
	}
	if req == nil || strings.TrimSpace(req.Message) == "" {
		return nil, domain.ErrEmptyMessage
	}
	return s.client.GenerateReply(ctx, req.Message, req.PhoneNumber)
}
