package presentation

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wa-serv/config"
	"github.com/wa-serv/internal/domain"
)

// AIHandler serves the AI reply-suggestion endpoint. It is always registered so
// clients get a clear 503 when the feature is disabled instead of a 404.
type AIHandler struct {
	aiService domain.AIService
	cfg       config.AIConfig
}

// NewAIHandler creates a new AI handler.
func NewAIHandler(aiService domain.AIService, cfg config.AIConfig) *AIHandler {
	return &AIHandler{aiService: aiService, cfg: cfg}
}

// GenerateAIReply handles POST /api/ai/reply.
func (h *AIHandler) GenerateAIReply(c *gin.Context) {
	if !h.cfg.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "AI response feature is disabled",
		})
		return
	}

	var req domain.AIReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "message is required",
		})
		return
	}

	resp, err := h.aiService.GenerateReply(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case domain.ErrAIResponseDisabled:
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"message": "AI response feature is disabled",
			})
		case domain.ErrEmptyMessage:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "message is required",
			})
		default:
			// Don't leak upstream details to the caller.
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "failed to generate AI reply",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}
