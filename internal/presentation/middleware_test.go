package presentation

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/wa-serv/internal/mocks"
)

func TestBasicAuthMiddleware_ValidCredentials(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}
	middleware := AuthMiddleware(mockAuthService)

	router := setupTestRouter()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	mockAuthService.On("ValidateCredentials", "testuser", "testpass").Return(true)

	// Prepare request with basic auth
	req, _ := http.NewRequest("GET", "/test", nil)
	auth := base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))
	req.Header.Set("Authorization", "Basic "+auth)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestBasicAuthMiddleware_InvalidCredentials(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}
	middleware := AuthMiddleware(mockAuthService)

	router := setupTestRouter()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	mockAuthService.On("ValidateCredentials", "testuser", "wrongpass").Return(false)

	// Prepare request with invalid basic auth
	req, _ := http.NewRequest("GET", "/test", nil)
	auth := base64.StdEncoding.EncodeToString([]byte("testuser:wrongpass"))
	req.Header.Set("Authorization", "Basic "+auth)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestBasicAuthMiddleware_NoAuthHeader(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}
	middleware := AuthMiddleware(mockAuthService)

	router := setupTestRouter()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Prepare request without auth header
	req, _ := http.NewRequest("GET", "/test", nil)

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "Basic realm=\"WhatsPoints API\"", w.Header().Get("WWW-Authenticate"))
}

func TestBasicAuthMiddleware_InvalidAuthFormat(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}
	middleware := AuthMiddleware(mockAuthService)

	router := setupTestRouter()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Prepare request with invalid auth format
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestBasicAuthMiddleware_InvalidBase64(t *testing.T) {
	// Arrange
	mockAuthService := &mocks.MockAuthService{}
	middleware := AuthMiddleware(mockAuthService)

	router := setupTestRouter()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Prepare request with invalid base64
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic invalid_base64!")

	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
