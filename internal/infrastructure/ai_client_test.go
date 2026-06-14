package infrastructure

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wa-serv/internal/domain"
)

func TestAIClient_GenerateReply_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/ai/reply", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var body map[string]any
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "promo masih ada?", body["customer_message"])
		assert.Equal(t, "628123", body["phone_number"])

		json.NewEncoder(w).Encode(domain.AIReplyResponse{
			Reply:  "Masih kak",
			Intent: "ask_promo",
			Sources: []domain.AISource{
				{ID: 1, Title: "Promo", Content: "isi", Category: "promo", Score: 0.12},
			},
		})
	}))
	defer server.Close()

	client := NewAIClient(server.URL)
	resp, err := client.GenerateReply(context.Background(), "promo masih ada?", "628123")

	assert.NoError(t, err)
	assert.Equal(t, "Masih kak", resp.Reply)
	assert.Equal(t, "ask_promo", resp.Intent)
	assert.Len(t, resp.Sources, 1)
	assert.Equal(t, int64(1), resp.Sources[0].ID)
}

func TestAIClient_GenerateReply_Non2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewAIClient(server.URL)
	resp, err := client.GenerateReply(context.Background(), "hi", "")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "500")
}
