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
