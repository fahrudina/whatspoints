package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_ValidateCredentials_Success(t *testing.T) {
	// Arrange
	service := NewAuthService("testuser", "testpass")

	// Act
	result := service.ValidateCredentials("testuser", "testpass")

	// Assert
	assert.True(t, result)
}

func TestAuthService_ValidateCredentials_WrongUsername(t *testing.T) {
	// Arrange
	service := NewAuthService("testuser", "testpass")

	// Act
	result := service.ValidateCredentials("wronguser", "testpass")

	// Assert
	assert.False(t, result)
}

func TestAuthService_ValidateCredentials_WrongPassword(t *testing.T) {
	// Arrange
	service := NewAuthService("testuser", "testpass")

	// Act
	result := service.ValidateCredentials("testuser", "wrongpass")

	// Assert
	assert.False(t, result)
}

func TestAuthService_ValidateCredentials_EmptyCredentials(t *testing.T) {
	// Arrange
	service := NewAuthService("testuser", "testpass")

	// Act
	result1 := service.ValidateCredentials("", "testpass")
	result2 := service.ValidateCredentials("testuser", "")
	result3 := service.ValidateCredentials("", "")

	// Assert
	assert.False(t, result1)
	assert.False(t, result2)
	assert.False(t, result3)
}

func TestAuthService_ValidateCredentials_BothWrong(t *testing.T) {
	// Arrange
	service := NewAuthService("testuser", "testpass")

	// Act
	result := service.ValidateCredentials("wronguser", "wrongpass")

	// Assert
	assert.False(t, result)
}
