package presentation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wa-serv/config"
	"github.com/wa-serv/internal/domain"
	"github.com/wa-serv/internal/mocks"
)

func postJSON(handler gin.HandlerFunc, body any) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/ai/reply", handler)

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/ai/reply", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestAIHandler_Disabled_Returns503AndDoesNotCallService(t *testing.T) {
	svc := &mocks.MockAIService{}
	handler := NewAIHandler(svc, config.AIConfig{Enabled: false})

	w := postJSON(handler.GenerateAIReply, domain.AIReplyRequest{Message: "promo?"})

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var resp map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "AI response feature is disabled", resp["message"])
	svc.AssertNotCalled(t, "GenerateReply", mock.Anything, mock.Anything)
}

func TestAIHandler_EmptyMessage_Returns400(t *testing.T) {
	svc := &mocks.MockAIService{}
	handler := NewAIHandler(svc, config.AIConfig{Enabled: true})

	w := postJSON(handler.GenerateAIReply, domain.AIReplyRequest{Message: "  "})

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "message is required", resp["message"])
	svc.AssertNotCalled(t, "GenerateReply", mock.Anything, mock.Anything)
}

func TestAIHandler_Success_Returns200(t *testing.T) {
	svc := &mocks.MockAIService{}
	expected := &domain.AIReplyResponse{
		Reply:   "Masih kak",
		Intent:  "ask_promo",
		Sources: []domain.AISource{{ID: 1, Title: "Promo", Content: "isi", Category: "promo", Score: 0.12}},
	}
	svc.On("GenerateReply", mock.Anything, mock.MatchedBy(func(r *domain.AIReplyRequest) bool {
		return r.Message == "promo?" && r.PhoneNumber == "628"
	})).Return(expected, nil)

	handler := NewAIHandler(svc, config.AIConfig{Enabled: true})
	w := postJSON(handler.GenerateAIReply, domain.AIReplyRequest{Message: "promo?", PhoneNumber: "628"})

	assert.Equal(t, http.StatusOK, w.Code)
	var resp domain.AIReplyResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Masih kak", resp.Reply)
	assert.Equal(t, "ask_promo", resp.Intent)
	assert.Len(t, resp.Sources, 1)
	svc.AssertExpectations(t)
}
