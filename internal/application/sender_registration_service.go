package application

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/wa-serv/internal/domain"
	"github.com/wa-serv/repository"
	"github.com/wa-serv/whatsapp"
	"go.mau.fi/whatsmeow"
	waCompanionReg "go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

// RegistrationSession tracks an ongoing registration
type RegistrationSession struct {
	SessionID   string
	Client      *whatsmeow.Client
	Status      string // pending, connected, failed
	SenderID    string
	QRCode      string
	PairingCode string
	PhoneNumber string
	CreatedAt   time.Time
	mu          sync.RWMutex
}

// SenderRegistrationService implements sender registration business logic
type SenderRegistrationService struct {
	db            *sql.DB
	clientManager *whatsapp.ClientManager
	sessions      map[string]*RegistrationSession
	sessionsMu    sync.RWMutex
}

// NewSenderRegistrationService creates a new sender registration service
func NewSenderRegistrationService(db *sql.DB, clientManager *whatsapp.ClientManager) *SenderRegistrationService {
	return &SenderRegistrationService{
		db:            db,
		clientManager: clientManager,
		sessions:      make(map[string]*RegistrationSession),
	}
}

// StartQRRegistration starts a new QR code registration session
func (s *SenderRegistrationService) StartQRRegistration(ctx context.Context) (*domain.RegisterSenderQRResponse, error) {
	sessionID := uuid.New().String()

	// Create a new device store for the new phone number
	deviceStore := s.clientManager.GetContainer().NewDevice()

	// Set custom device name and platform type before pairing
	store.DeviceProps.Os = proto.String(whatsapp.DeviceName)
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_DESKTOP.Enum()

	logLevel := whatsapp.GetLogLevel()
	clientLog := waLog.Stdout("RegisterSession", logLevel, true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Disable history sync to save bandwidth and resources
	// We only need to send messages, not receive history
	client.EnableAutoReconnect = true
	client.AutomaticMessageRerequestFromPhone = false

	// Create session
	session := &RegistrationSession{
		SessionID: sessionID,
		Client:    client,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Add event handler to track connection status
	client.AddEventHandler(func(evt interface{}) {
		// First, handle registration-specific events
		switch evt.(type) {
		case *events.PairSuccess:
			session.mu.Lock()
			session.Status = "connected"
			if client.Store.ID != nil {
				session.SenderID = client.Store.ID.User
				// Register sender in database
				s.registerSender(session.SenderID, client.Store.ID.User)
			}
			session.mu.Unlock()
		case *events.LoggedOut:
			session.mu.Lock()
			session.Status = "failed"
			session.mu.Unlock()
		case *events.Connected:
			// Client connected to WhatsApp servers
		case *events.Disconnected:
			// Only mark as failed if not already connected
			session.mu.Lock()
			if session.Status == "pending" {
				session.Status = "failed"
			}
			session.mu.Unlock()
		}

		// Then, let whatsapp package handle all events normally
		// This ensures proper WebSocket connection maintenance
		whatsapp.HandleEvent(evt, s.db, client)
	})

	// Get QR code channel - use background context to keep it alive beyond the HTTP request
	// DO NOT use the HTTP request context here, as it will cancel when the request ends
	qrCtx, cancelQR := context.WithCancel(context.Background())
	qrChan, err := client.GetQRChannel(qrCtx)
	if err != nil {
		return &domain.RegisterSenderQRResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get QR channel: %v", err),
		}, err
	}

	// Connect the client
	if err := client.Connect(); err != nil {
		return &domain.RegisterSenderQRResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to connect: %v", err),
		}, err
	}

	// Wait for the first QR code and convert to base64 image
	// Keep this goroutine running to handle QR refreshes
	firstQRReceived := make(chan bool, 1)
	go func() {
		defer fmt.Println("QR goroutine exiting")
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Printf("QR Code received from WhatsApp, raw code length: %d\n", len(evt.Code))
				// Generate QR code as PNG image
				qrBytes := generateQRCodePNG(evt.Code)
				if len(qrBytes) == 0 {
					fmt.Println("Error: Failed to generate QR code PNG")
					continue
				}

				qrBase64 := base64.StdEncoding.EncodeToString(qrBytes)

				session.mu.Lock()
				session.QRCode = qrBase64
				session.mu.Unlock()

				fmt.Printf("QR Code PNG generated and stored (base64 length: %d bytes)\n", len(qrBase64))

				// Signal that first QR is ready
				select {
				case firstQRReceived <- true:
					fmt.Println("First QR code signaled")
				default:
					fmt.Println("QR code updated (not first)")
					// Subsequent QR codes don't need to signal
				}
			} else if evt.Event == "success" {
				fmt.Println("QR Code scan successful!")
				session.mu.Lock()
				session.Status = "connected"
				session.mu.Unlock()
				// Don't break here - let the channel close naturally
				fmt.Println("Waiting for pairing to complete...")
			} else {
				fmt.Printf("QR Event: %s\n", evt.Event)
			}
		}
		fmt.Println("QR Channel closed")
	}()

	// Wait for the first QR code with timeout
	select {
	case <-firstQRReceived:
		// QR code ready
		fmt.Println("QR code ready to be returned to client")
	case <-time.After(10 * time.Second):
		// Timeout waiting for QR code - increased from 5 to 10 seconds
		fmt.Println("Timeout waiting for QR code")
		cancelQR()
		client.Disconnect()
		return &domain.RegisterSenderQRResponse{
			Success: false,
			Message: "Timeout waiting for QR code generation",
		}, fmt.Errorf("timeout waiting for QR code")
	}

	// Store session
	s.sessionsMu.Lock()
	s.sessions[sessionID] = session
	s.sessionsMu.Unlock()

	// Clean up old sessions (older than 10 minutes)
	go s.cleanupOldSessions()

	// Get the QR code from session
	session.mu.RLock()
	qrCode := session.QRCode
	session.mu.RUnlock()

	fmt.Printf("Returning QR code response (base64 length: %d)\n", len(qrCode))

	return &domain.RegisterSenderQRResponse{
		Success:   true,
		SessionID: sessionID,
		QRCode:    qrCode,
		Message:   "QR code generated. Please scan with WhatsApp.",
	}, nil
}

// StartCodeRegistration starts a new pairing code registration session
func (s *SenderRegistrationService) StartCodeRegistration(ctx context.Context, req *domain.RegisterSenderCodeRequest) (*domain.RegisterSenderCodeResponse, error) {
	if req.PhoneNumber == "" {
		return &domain.RegisterSenderCodeResponse{
			Success: false,
			Message: "Phone number is required",
		}, fmt.Errorf("phone number is required")
	}

	sessionID := uuid.New().String()

	// Clean phone number (remove +, spaces, etc.)
	cleanedPhone := cleanPhoneNumber(req.PhoneNumber)

	// Create a new device store for the new phone number
	deviceStore := s.clientManager.GetContainer().NewDevice()

	// Set custom device name and platform type before pairing
	store.DeviceProps.Os = proto.String(whatsapp.DeviceName)
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_DESKTOP.Enum()

	logLevel := whatsapp.GetLogLevel()
	clientLog := waLog.Stdout("RegisterSession", logLevel, true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// Disable history sync to save bandwidth and resources
	// We only need to send messages, not receive history
	client.EnableAutoReconnect = true
	client.AutomaticMessageRerequestFromPhone = false

	// Create session
	session := &RegistrationSession{
		SessionID:   sessionID,
		Client:      client,
		Status:      "pending",
		PhoneNumber: cleanedPhone,
		CreatedAt:   time.Now(),
	}

	// Add event handler to track connection status
	client.AddEventHandler(func(evt interface{}) {
		// First, handle registration-specific events
		switch evt.(type) {
		case *events.PairSuccess:
			session.mu.Lock()
			session.Status = "connected"
			if client.Store.ID != nil {
				session.SenderID = client.Store.ID.User
				// Register sender in database
				s.registerSender(session.SenderID, cleanedPhone)
			}
			session.mu.Unlock()
		case *events.LoggedOut:
			session.mu.Lock()
			session.Status = "failed"
			session.mu.Unlock()
		}

		// Then, let whatsapp package handle all events normally
		// This ensures proper WebSocket connection maintenance
		whatsapp.HandleEvent(evt, s.db, client)
	})

	// Connect first
	if err := client.Connect(); err != nil {
		return &domain.RegisterSenderCodeResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to connect: %v", err),
		}, err
	}

	// Request pairing code
	code, err := client.PairPhone(ctx, cleanedPhone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		client.Disconnect()
		return &domain.RegisterSenderCodeResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to request pairing code: %v", err),
		}, err
	}

	session.PairingCode = code

	// Store session
	s.sessionsMu.Lock()
	s.sessions[sessionID] = session
	s.sessionsMu.Unlock()

	// Clean up old sessions
	go s.cleanupOldSessions()

	return &domain.RegisterSenderCodeResponse{
		Success:     true,
		SessionID:   sessionID,
		PairingCode: code,
		PhoneNumber: req.PhoneNumber,
		Message:     "Pairing code generated. Please enter it in WhatsApp.",
	}, nil
}

// GetRegistrationStatus returns the status of a registration session
func (s *SenderRegistrationService) GetRegistrationStatus(ctx context.Context, sessionID string) (*domain.RegistrationStatusResponse, error) {
	s.sessionsMu.RLock()
	session, exists := s.sessions[sessionID]
	s.sessionsMu.RUnlock()

	if !exists {
		return &domain.RegistrationStatusResponse{
			Success: false,
			Status:  "not_found",
			Message: "Registration session not found or expired",
		}, nil
	}

	session.mu.RLock()
	status := session.Status
	senderID := session.SenderID
	qrCode := session.QRCode
	session.mu.RUnlock()

	response := &domain.RegistrationStatusResponse{
		Success:  true,
		Status:   status,
		SenderID: senderID,
	}

	switch status {
	case "pending":
		response.Message = "Waiting for WhatsApp pairing..."
	case "connected":
		response.Message = "Successfully registered! Sender ID: " + senderID
		// Add client to manager if successful
		if senderID != "" {
			s.clientManager.AddExistingClient(session.Client, senderID)
		}
		// Clean up this session after successful registration
		s.sessionsMu.Lock()
		delete(s.sessions, sessionID)
		s.sessionsMu.Unlock()
	case "failed":
		response.Message = "Registration failed. Please try again."
		// Clean up failed session
		s.sessionsMu.Lock()
		if session.Client != nil {
			session.Client.Disconnect()
		}
		delete(s.sessions, sessionID)
		s.sessionsMu.Unlock()
	}

	// Include updated QR code for pending registrations (QR codes can refresh)
	if status == "pending" && qrCode != "" {
		// QR codes expire and refresh, so we need to send the latest one
		response.QRCode = qrCode
	}

	return response, nil
}

// registerSender creates a sender record in the database
func (s *SenderRegistrationService) registerSender(senderID, phoneNumber string) {
	name := fmt.Sprintf("Sender %s", senderID)

	// Check if this is the first sender (make it default)
	senders, err := repository.GetAllSenders(s.db)
	isDefault := err != nil || len(senders) == 0

	err = repository.CreateSenderIfNotExists(s.db, senderID, phoneNumber, name, isDefault)
	if err != nil {
		fmt.Printf("Failed to create sender record: %v\n", err)
	}
}

// cleanupOldSessions removes sessions older than 10 minutes
func (s *SenderRegistrationService) cleanupOldSessions() {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	for sessionID, session := range s.sessions {
		if session.CreatedAt.Before(cutoff) {
			if session.Client != nil {
				session.Client.Disconnect()
			}
			delete(s.sessions, sessionID)
		}
	}
}

// cleanPhoneNumber removes non-digit characters
func cleanPhoneNumber(phone string) string {
	cleaned := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			cleaned += string(char)
		}
	}
	return cleaned
}

// generateQRCodePNG generates a QR code as PNG bytes
func generateQRCodePNG(code string) []byte {
	// Generate QR code as PNG image with higher error correction for better scanning
	// Use High error correction level and larger size for WhatsApp QR codes
	png, err := qrcode.Encode(code, qrcode.High, 512)
	if err != nil {
		fmt.Printf("Failed to generate QR code: %v\n", err)
		return []byte{}
	}
	fmt.Printf("Generated QR code PNG: %d bytes for code: %s\n", len(png), code[:20]+"...")
	return png
}
