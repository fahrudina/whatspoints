package application

import "github.com/wa-serv/internal/domain"

type authService struct {
	username string
	password string
}

// NewAuthService creates a new auth service
func NewAuthService(username, password string) domain.AuthService {
	return &authService{
		username: username,
		password: password,
	}
}

// ValidateCredentials validates the provided credentials
func (s *authService) ValidateCredentials(username, password string) bool {
	return s.username == username && s.password == password
}
