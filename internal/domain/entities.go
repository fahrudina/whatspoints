package domain

// Message represents a WhatsApp message
type Message struct {
	ID      string
	To      string
	Content string
	SentAt  string
}

// SendMessageRequest represents the request to send a message
type SendMessageRequest struct {
	To      string `json:"to" validate:"required"`
	Message string `json:"message" validate:"required"`
	From    string `json:"from,omitempty"` // Optional: sender phone number identifier
}

// SendMessageResponse represents the response after sending a message
type SendMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

// WhatsAppStatus represents the status of WhatsApp client
type WhatsAppStatus struct {
	Connected bool   `json:"connected"`
	LoggedIn  bool   `json:"logged_in"`
	JID       string `json:"jid,omitempty"`
}

// ServiceStatus represents the overall service status
type ServiceStatus struct {
	WhatsApp WhatsAppStatus `json:"whatsapp"`
}

// Sender represents a WhatsApp sender account
type Sender struct {
	ID          string `json:"id"`           // Unique identifier for the sender
	PhoneNumber string `json:"phone_number"` // Phone number in WhatsApp format
	Name        string `json:"name"`         // Friendly name for the sender
	IsDefault   bool   `json:"is_default"`   // Whether this is the default sender
	IsActive    bool   `json:"is_active"`    // Whether this sender is currently active
}

// RegisterSenderQRRequest represents the request to start QR registration
type RegisterSenderQRRequest struct {
	SessionID string `json:"session_id,omitempty"` // Optional session ID for tracking
}

// RegisterSenderQRResponse represents the response for QR registration
type RegisterSenderQRResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id"`          // Session ID for status checking
	QRCode    string `json:"qr_code,omitempty"`   // Base64 encoded QR code image
	Message   string `json:"message,omitempty"`   // Status or error message
}

// RegisterSenderCodeRequest represents the request to register with pairing code
type RegisterSenderCodeRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"` // Phone number with country code
}

// RegisterSenderCodeResponse represents the response for code registration
type RegisterSenderCodeResponse struct {
	Success      bool   `json:"success"`
	SessionID    string `json:"session_id"`          // Session ID for status checking
	PairingCode  string `json:"pairing_code,omitempty"` // The pairing code to enter in WhatsApp
	PhoneNumber  string `json:"phone_number,omitempty"` // Phone number being registered
	Message      string `json:"message,omitempty"`   // Status or error message
}

// RegistrationStatusResponse represents the status of a registration session
type RegistrationStatusResponse struct {
	Success   bool   `json:"success"`
	Status    string `json:"status"`              // pending, connected, failed
	SenderID  string `json:"sender_id,omitempty"` // Set when successfully connected
	Message   string `json:"message,omitempty"`   // Status or error message
}
