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
	"github.com/wa-serv/internal/domain"
	"github.com/wa-serv/internal/mocks"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestMessageHandler_SendMessage_Success(t *testing.T) {
	// Arrange
	mockMessageService := &mocks.MockMessageService{}
	mockAuthService := &mocks.MockAuthService{}
	handler := NewMessageHandler(mockMessageService, mockAuthService)

	router := setupTestRouter()
	router.POST("/message", handler.SendMessage)

	reqBody := domain.SendMessageRequest{
		To:      "+1234567890",
		Message: "Test message",
	}
	expectedResponse := &domain.SendMessageResponse{
		Success: true,
		Message: "Message sent successfully",
		ID:      "test-id",
	}

	mockMessageService.On("SendMessage", mock.Anything, &reqBody).Return(expectedResponse, nil)

	// Prepare request
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/message", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.SendMessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Message sent successfully", response.Message)
	assert.Equal(t, "test-id", response.ID)

	mockMessageService.AssertExpectations(t)
}

func TestMessageHandler_SendMessage_InvalidJSON(t *testing.T) {
	// Arrange
	mockMessageService := &mocks.MockMessageService{}
	mockAuthService := &mocks.MockAuthService{}
	handler := NewMessageHandler(mockMessageService, mockAuthService)

	router := setupTestRouter()
	router.POST("/message", handler.SendMessage)

	// Prepare invalid JSON request
	req, _ := http.NewRequest("POST", "/message", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response domain.SendMessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
}

func TestMessageHandler_SendMessage_ServiceError(t *testing.T) {
	// Arrange
	mockMessageService := &mocks.MockMessageService{}
	mockAuthService := &mocks.MockAuthService{}
	handler := NewMessageHandler(mockMessageService, mockAuthService)

	router := setupTestRouter()
	router.POST("/message", handler.SendMessage)

	reqBody := domain.SendMessageRequest{
		To:      "123", // Invalid phone number
		Message: "Test message",
	}
	expectedResponse := &domain.SendMessageResponse{
		Success: false,
		Message: "Invalid phone number format",
	}

	mockMessageService.On("SendMessage", mock.Anything, &reqBody).Return(expectedResponse, domain.ErrInvalidPhoneNumber)

	// Prepare request
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/message", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response domain.SendMessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Invalid phone number format", response.Message)

	mockMessageService.AssertExpectations(t)
}

func TestMessageHandler_GetStatus_Success(t *testing.T) {
	// Arrange
	mockMessageService := &mocks.MockMessageService{}
	mockAuthService := &mocks.MockAuthService{}
	handler := NewMessageHandler(mockMessageService, mockAuthService)

	router := setupTestRouter()
	router.GET("/status", handler.GetStatus)

	expectedStatus := &domain.ServiceStatus{
		WhatsApp: domain.WhatsAppStatus{
			Connected: true,
			LoggedIn:  true,
			JID:       "test@s.whatsapp.net",
		},
	}

	mockMessageService.On("GetStatus", mock.Anything).Return(expectedStatus, nil)

	// Prepare request
	req, _ := http.NewRequest("GET", "/status", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.ServiceStatus
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.WhatsApp.Connected)
	assert.True(t, response.WhatsApp.LoggedIn)
	assert.Equal(t, "test@s.whatsapp.net", response.WhatsApp.JID)

	mockMessageService.AssertExpectations(t)
}

func TestMessageHandler_GetStatus_ServiceError(t *testing.T) {
	// Arrange
	mockMessageService := &mocks.MockMessageService{}
	mockAuthService := &mocks.MockAuthService{}
	handler := NewMessageHandler(mockMessageService, mockAuthService)

	router := setupTestRouter()
	router.GET("/status", handler.GetStatus)

	mockMessageService.On("GetStatus", mock.Anything).Return((*domain.ServiceStatus)(nil), domain.ErrWhatsAppNotConnected)

	// Prepare request
	req, _ := http.NewRequest("GET", "/status", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "whatsapp client is not connected", response["error"])

	mockMessageService.AssertExpectations(t)
}

func TestMessageHandler_ListSenders_Success(t *testing.T) {
	// Arrange
	mockMessageService := &mocks.MockMessageService{}
	mockAuthService := &mocks.MockAuthService{}
	handler := NewMessageHandler(mockMessageService, mockAuthService)

	router := setupTestRouter()
	router.GET("/senders", handler.ListSenders)

	expectedSenders := []*domain.Sender{
		{
			ID:          "1234567890",
			PhoneNumber: "1234567890",
			Name:        "Sender 1234567890",
			IsDefault:   true,
			IsActive:    true,
		},
		{
			ID:          "9876543210",
			PhoneNumber: "9876543210",
			Name:        "Sender 9876543210",
			IsDefault:   false,
			IsActive:    true,
		},
	}

	mockMessageService.On("ListSenders", mock.Anything).Return(expectedSenders, nil)

	// Prepare request
	req, _ := http.NewRequest("GET", "/senders", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["count"])

	senders := response["senders"].([]interface{})
	assert.Len(t, senders, 2)

	firstSender := senders[0].(map[string]interface{})
	assert.Equal(t, "1234567890", firstSender["id"])
	assert.Equal(t, true, firstSender["is_default"])

	mockMessageService.AssertExpectations(t)
}

func TestMessageHandler_ListSenders_ServiceError(t *testing.T) {
	// Arrange
	mockMessageService := &mocks.MockMessageService{}
	mockAuthService := &mocks.MockAuthService{}
	handler := NewMessageHandler(mockMessageService, mockAuthService)

	router := setupTestRouter()
	router.GET("/senders", handler.ListSenders)

	mockMessageService.On("ListSenders", mock.Anything).Return(([]*domain.Sender)(nil), domain.ErrNoActiveSender)

	// Prepare request
	req, _ := http.NewRequest("GET", "/senders", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "no active sender available", response["error"])

	mockMessageService.AssertExpectations(t)
}
