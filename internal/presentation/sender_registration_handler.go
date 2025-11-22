package presentation

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wa-serv/internal/domain"
)

type SenderRegistrationHandler struct {
	registrationService domain.SenderRegistrationService
	authService         domain.AuthService
}

// NewSenderRegistrationHandler creates a new sender registration handler
func NewSenderRegistrationHandler(registrationService domain.SenderRegistrationService, authService domain.AuthService) *SenderRegistrationHandler {
	return &SenderRegistrationHandler{
		registrationService: registrationService,
		authService:         authService,
	}
}

// StartQRRegistration handles POST /api/register-sender-qr
func (h *SenderRegistrationHandler) StartQRRegistration(c *gin.Context) {
	response, err := h.registrationService.StartQRRegistration(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// StartCodeRegistration handles POST /api/register-sender-code
func (h *SenderRegistrationHandler) StartCodeRegistration(c *gin.Context) {
	var req domain.RegisterSenderCodeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.RegisterSenderCodeResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	response, err := h.registrationService.StartCodeRegistration(c.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if response != nil && !response.Success {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetRegistrationStatus handles GET /api/register-sender-status/:sessionId
func (h *SenderRegistrationHandler) GetRegistrationStatus(c *gin.Context) {
	sessionID := c.Param("sessionId")
	
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, domain.RegistrationStatusResponse{
			Success: false,
			Status:  "error",
			Message: "Session ID is required",
		})
		return
	}

	response, err := h.registrationService.GetRegistrationStatus(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.RegistrationStatusResponse{
			Success: false,
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
