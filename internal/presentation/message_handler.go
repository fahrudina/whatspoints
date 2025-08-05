package presentation

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wa-serv/internal/domain"
)

type MessageHandler struct {
	messageService domain.MessageService
	authService    domain.AuthService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService domain.MessageService, authService domain.AuthService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		authService:    authService,
	}
}

// SendMessage handles POST /api/send-message
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req domain.SendMessageRequest

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.SendMessageResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Send message using service
	response, err := h.messageService.SendMessage(c.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError

		// Map domain errors to HTTP status codes
		switch err {
		case domain.ErrWhatsAppNotConnected:
			statusCode = http.StatusServiceUnavailable
		case domain.ErrInvalidPhoneNumber:
			statusCode = http.StatusBadRequest
		case domain.ErrMessageSendFailed:
			statusCode = http.StatusInternalServerError
		}

		c.JSON(statusCode, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetStatus handles GET /api/status
func (h *MessageHandler) GetStatus(c *gin.Context) {
	status, err := h.messageService.GetStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// HealthCheck handles GET /health
func (h *MessageHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "whatspoints-api",
	})
}
