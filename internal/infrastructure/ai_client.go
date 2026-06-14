package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wa-serv/internal/domain"
)

// AIClient is an HTTP client for the Python AI sidecar service.
type AIClient struct {
	baseURL string
	client  *http.Client
}

// aiReplyRequest is the wire payload expected by the sidecar's POST /ai/reply.
type aiReplyRequest struct {
	CustomerMessage string `json:"customer_message"`
	PhoneNumber     string `json:"phone_number,omitempty"`
}

// NewAIClient creates an AI sidecar client with a 15s request timeout.
func NewAIClient(baseURL string) *AIClient {
	return &AIClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// GenerateReply calls POST {baseURL}/ai/reply and returns the suggested reply.
func (a *AIClient) GenerateReply(ctx context.Context, message, phoneNumber string) (*domain.AIReplyResponse, error) {
	body, err := json.Marshal(aiReplyRequest{CustomerMessage: message, PhoneNumber: phoneNumber})
	if err != nil {
		return nil, fmt.Errorf("encode ai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/ai/reply", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build ai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call ai service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ai service returned status %d", resp.StatusCode)
	}

	var out domain.AIReplyResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode ai response: %w", err)
	}
	return &out, nil
}
