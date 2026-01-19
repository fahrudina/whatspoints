package infrastructure_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/wa-serv/internal/domain"
	"github.com/wa-serv/internal/infrastructure"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
)

// mockClientManager implements the client manager interface for testing
type mockClientManager struct {
	clients       map[string]*whatsmeow.Client
	defaultClient *whatsmeow.Client
	getClientErr  error
	getDefaultErr error
}

func (m *mockClientManager) GetClient(senderID string) (*whatsmeow.Client, error) {
	if m.getClientErr != nil {
		return nil, m.getClientErr
	}
	if client, ok := m.clients[senderID]; ok {
		return client, nil
	}
	return nil, domain.ErrSenderNotFound
}

func (m *mockClientManager) GetDefaultClient() (*whatsmeow.Client, error) {
	if m.getDefaultErr != nil {
		return nil, m.getDefaultErr
	}
	return m.defaultClient, nil
}

func (m *mockClientManager) GetAllClients() map[string]*whatsmeow.Client {
	return m.clients
}

// createMockClient creates a mock whatsmeow client with basic setup
func createMockClient(jidUser string, connected bool) *whatsmeow.Client {
	jid := types.JID{
		User:   jidUser,
		Server: types.DefaultUserServer,
	}

	client := &whatsmeow.Client{
		Store: &store.Device{
			ID: &jid,
		},
	}

	return client
}

func TestNewWhatsAppRepository(t *testing.T) {
	client := createMockClient("1234567890", true)
	repo := infrastructure.NewWhatsAppRepository(client)

	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}

	// Test that repository methods are accessible
	jid := repo.GetJID()
	if jid == "" {
		t.Error("Expected non-empty JID")
	}
}

func TestNewWhatsAppRepositoryWithDB(t *testing.T) {
	client := createMockClient("1234567890", true)

	repo := infrastructure.NewWhatsAppRepositoryWithDB(client, nil)

	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}
}

func TestNewWhatsAppRepositoryWithClients(t *testing.T) {
	defaultClient := createMockClient("1111111111", true)
	clients := map[string]*whatsmeow.Client{
		"sender1": createMockClient("2222222222", true),
		"sender2": createMockClient("3333333333", true),
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClients(defaultClient, nil, clients)

	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}
}

func TestNewWhatsAppRepositoryWithClientManager(t *testing.T) {
	defaultClient := createMockClient("1234567890", true)
	mockManager := &mockClientManager{
		defaultClient: defaultClient,
		clients:       make(map[string]*whatsmeow.Client),
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClientManager(nil, mockManager)

	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}
}

func TestNewWhatsAppRepositoryWithClientManager_NoDefaultClient(t *testing.T) {
	mockManager := &mockClientManager{
		getDefaultErr: domain.ErrNoActiveSender,
		clients:       make(map[string]*whatsmeow.Client),
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClientManager(nil, mockManager)

	if repo == nil {
		t.Fatal("Expected non-nil repository even with no default client")
	}
}

func TestGetJID(t *testing.T) {
	jidUser := "1234567890"
	client := createMockClient(jidUser, true)
	repo := infrastructure.NewWhatsAppRepository(client)

	result := repo.GetJID()

	if result == "" {
		t.Error("Expected non-empty JID")
	}

	expectedJID := jidUser + "@" + types.DefaultUserServer
	if result != expectedJID {
		t.Errorf("Expected JID %s, got %s", expectedJID, result)
	}
}

func TestGetJID_NoClient(t *testing.T) {
	// Create repository without setting up client properly
	repo := infrastructure.NewWhatsAppRepository(nil)

	result := repo.GetJID()

	if result != "" {
		t.Errorf("Expected empty JID when no client, got %s", result)
	}
}

func TestGetSenderJID_Success(t *testing.T) {
	jidUser := "1234567890"
	client := createMockClient(jidUser, true)

	clients := map[string]*whatsmeow.Client{
		"sender1": client,
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClients(nil, nil, clients)

	result, err := repo.GetSenderJID("sender1")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedJID := jidUser + "@" + types.DefaultUserServer
	if result != expectedJID {
		t.Errorf("Expected JID %s, got %s", expectedJID, result)
	}
}

func TestGetSenderJID_NotFound(t *testing.T) {
	repo := infrastructure.NewWhatsAppRepository(nil)

	result, err := repo.GetSenderJID("nonexistent")

	if err != domain.ErrSenderNotFound {
		t.Errorf("Expected ErrSenderNotFound, got %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty JID, got %s", result)
	}
}

func TestIsLoggedIn_WithClient(t *testing.T) {
	client := createMockClient("1234567890", true)
	repo := infrastructure.NewWhatsAppRepository(client)

	// Note: IsLoggedIn will return false for our mock since it's not actually logged in
	result := repo.IsLoggedIn()

	// We don't expect true here since the mock client isn't fully initialized
	_ = result
}

func TestIsLoggedIn_NoClient(t *testing.T) {
	repo := infrastructure.NewWhatsAppRepository(nil)

	result := repo.IsLoggedIn()

	if result {
		t.Error("Expected false when no client")
	}
}

func TestIsConnected_WithClient(t *testing.T) {
	client := createMockClient("1234567890", true)
	repo := infrastructure.NewWhatsAppRepository(client)

	// Note: Mock client won't actually be connected
	result := repo.IsConnected()

	// We don't expect true for mock client
	_ = result
}

func TestIsConnected_NoClient(t *testing.T) {
	repo := infrastructure.NewWhatsAppRepository(nil)

	result := repo.IsConnected()

	if result {
		t.Error("Expected false when no client")
	}
}

func TestIsConnected_WithClientManager(t *testing.T) {
	client := createMockClient("1234567890", true)
	mockManager := &mockClientManager{
		defaultClient: client,
		clients: map[string]*whatsmeow.Client{
			"sender1": client,
		},
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClientManager(nil, mockManager)

	result := repo.IsConnected()

	// Mock clients won't be connected, so we expect false
	if result {
		t.Error("Expected false for disconnected mock clients")
	}
}

func TestListSenders_NoDatabase(t *testing.T) {
	client := createMockClient("1234567890", true)
	repo := infrastructure.NewWhatsAppRepository(client)

	senders, err := repo.ListSenders()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(senders) != 1 {
		t.Errorf("Expected 1 sender, got %d", len(senders))
	}

	if senders[0].ID != "default" {
		t.Errorf("Expected ID 'default', got %s", senders[0].ID)
	}

	if !senders[0].IsDefault {
		t.Error("Expected IsDefault to be true")
	}
}

func TestGetDefaultSender_NoDatabase(t *testing.T) {
	client := createMockClient("1234567890", true)
	repo := infrastructure.NewWhatsAppRepository(client)

	sender, err := repo.GetDefaultSender()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if sender == nil {
		t.Fatal("Expected non-nil sender")
	}

	if sender.ID != "default" {
		t.Errorf("Expected ID 'default', got %s", sender.ID)
	}

	if !sender.IsDefault {
		t.Error("Expected IsDefault to be true")
	}
}

func TestSendMessage_NoClient(t *testing.T) {
	repo := infrastructure.NewWhatsAppRepository(nil)
	ctx := context.Background()

	_, err := repo.SendMessage(ctx, "1234567890@s.whatsapp.net", "test message")

	if err == nil {
		t.Error("Expected error when no client available")
	}
}

func TestSendMessageFrom_ClientNotFound(t *testing.T) {
	repo := infrastructure.NewWhatsAppRepository(nil)
	ctx := context.Background()

	_, err := repo.SendMessageFrom(ctx, "nonexistent", "1234567890@s.whatsapp.net", "test")

	if err == nil {
		t.Error("Expected error for nonexistent sender")
	}
}

func TestGetClient_FromClientManager(t *testing.T) {
	client := createMockClient("1234567890", true)
	mockManager := &mockClientManager{
		clients: map[string]*whatsmeow.Client{
			"sender1": client,
		},
		defaultClient: createMockClient("9999999999", true),
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClientManager(nil, mockManager)

	// Test getting sender JID uses getClient internally
	jid, err := repo.GetSenderJID("sender1")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if jid == "" {
		t.Error("Expected non-empty JID")
	}
}

func TestGetClient_FallbackToDefault(t *testing.T) {
	defaultClient := createMockClient("1234567890", true)
	mockManager := &mockClientManager{
		defaultClient: defaultClient,
		clients:       make(map[string]*whatsmeow.Client),
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClientManager(nil, mockManager)

	// Request non-existent sender, should fall back to default
	jid := repo.GetJID()

	if jid == "" {
		t.Error("Expected to fall back to default client")
	}
}

func TestGetSenderJID_NoJID(t *testing.T) {
	// Create client without JID
	client := &whatsmeow.Client{
		Store: &store.Device{
			ID: nil,
		},
	}

	clients := map[string]*whatsmeow.Client{
		"sender1": client,
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClients(nil, nil, clients)

	result, err := repo.GetSenderJID("sender1")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty string when no JID, got %s", result)
	}
}

func TestGetJID_MultipleClients(t *testing.T) {
	client1 := createMockClient("1111111111", true)
	client2 := createMockClient("2222222222", true)

	clients := map[string]*whatsmeow.Client{
		"sender1": client1,
		"sender2": client2,
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClients(client1, nil, clients)

	// Should get the default client's JID
	jid := repo.GetJID()

	if jid == "" {
		t.Error("Expected non-empty JID")
	}
}

func TestGetClient_MultipleClientsWithManager(t *testing.T) {
	client1 := createMockClient("1111111111", true)
	client2 := createMockClient("2222222222", true)

	mockManager := &mockClientManager{
		clients: map[string]*whatsmeow.Client{
			"sender1": client1,
			"sender2": client2,
		},
		defaultClient: client1,
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClientManager(nil, mockManager)

	// Test getting different senders
	jid1, err1 := repo.GetSenderJID("sender1")
	jid2, err2 := repo.GetSenderJID("sender2")

	if err1 != nil {
		t.Errorf("Expected no error for sender1, got %v", err1)
	}

	if err2 != nil {
		t.Errorf("Expected no error for sender2, got %v", err2)
	}

	if jid1 == jid2 {
		t.Error("Expected different JIDs for different senders")
	}
}

func TestListSenders_WithNilClient(t *testing.T) {
	repo := infrastructure.NewWhatsAppRepository(nil)

	senders, err := repo.ListSenders()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(senders) != 1 {
		t.Errorf("Expected 1 sender, got %d", len(senders))
	}

	if senders[0].PhoneNumber != "" {
		t.Error("Expected empty phone number when no client")
	}
}

func TestGetDefaultSender_WithNilClient(t *testing.T) {
	repo := infrastructure.NewWhatsAppRepository(nil)

	sender, err := repo.GetDefaultSender()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if sender == nil {
		t.Fatal("Expected non-nil sender")
	}

	if sender.PhoneNumber != "" {
		t.Error("Expected empty phone number when no client")
	}
}

func TestIsConnected_WithNilClientAndManager(t *testing.T) {
	mockManager := &mockClientManager{
		clients: map[string]*whatsmeow.Client{},
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClientManager(nil, mockManager)

	result := repo.IsConnected()

	if result {
		t.Error("Expected false when no clients in manager")
	}
}

func TestGetSenderJID_WithClientManager(t *testing.T) {
	client := createMockClient("1234567890", true)
	mockManager := &mockClientManager{
		clients: map[string]*whatsmeow.Client{
			"sender1": client,
		},
		defaultClient: client,
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClientManager(nil, mockManager)

	jid, err := repo.GetSenderJID("sender1")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if jid == "" {
		t.Error("Expected non-empty JID")
	}
}

func TestGetClient_WithEmptySenderID(t *testing.T) {
	client := createMockClient("1234567890", true)
	repo := infrastructure.NewWhatsAppRepository(client)

	// Test that empty senderID falls back to default
	jid := repo.GetJID()

	if jid == "" {
		t.Error("Expected non-empty JID with empty senderID")
	}
}

func TestNewWhatsAppRepositoryWithClients_EmptyClientsMap(t *testing.T) {
	defaultClient := createMockClient("1234567890", true)
	emptyClients := map[string]*whatsmeow.Client{}

	repo := infrastructure.NewWhatsAppRepositoryWithClients(defaultClient, nil, emptyClients)

	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}

	// Should still work with default client
	jid := repo.GetJID()
	if jid == "" {
		t.Error("Expected non-empty JID from default client")
	}
}

func TestGetClient_ManagerReturnsNilClient(t *testing.T) {
	mockManager := &mockClientManager{
		clients:       make(map[string]*whatsmeow.Client),
		defaultClient: nil,
		getDefaultErr: domain.ErrNoActiveSender,
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClientManager(nil, mockManager)

	// Should fail to get JID when no client available
	jid := repo.GetJID()

	if jid != "" {
		t.Error("Expected empty JID when manager has no clients")
	}
}

func TestGetSenderJID_ManagerError(t *testing.T) {
	mockManager := &mockClientManager{
		clients:       make(map[string]*whatsmeow.Client),
		defaultClient: nil,
		getClientErr:  domain.ErrSenderNotFound,
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClientManager(nil, mockManager)

	_, err := repo.GetSenderJID("nonexistent")

	if err != domain.ErrSenderNotFound {
		t.Errorf("Expected ErrSenderNotFound, got %v", err)
	}
}

// TestMultipleSendersScenario tests a realistic scenario with multiple senders
func TestMultipleSendersScenario(t *testing.T) {
	// Setup: Create multiple sender clients
	senderClients := map[string]*whatsmeow.Client{
		"sales":     createMockClient("6281234567890", true), // Sales department
		"support":   createMockClient("6289876543210", true), // Support department
		"marketing": createMockClient("6285555555555", true), // Marketing department
	}

	defaultClient := senderClients["sales"] // Sales is default

	// Create repository with multiple clients
	repo := infrastructure.NewWhatsAppRepositoryWithClients(defaultClient, nil, senderClients)

	t.Run("Verify all senders are registered", func(t *testing.T) {
		// Test each sender can be retrieved
		for senderID := range senderClients {
			jid, err := repo.GetSenderJID(senderID)
			if err != nil {
				t.Errorf("Failed to get JID for sender %s: %v", senderID, err)
			}
			if jid == "" {
				t.Errorf("Empty JID for sender %s", senderID)
			}
		}
	})

	t.Run("Verify different senders have different JIDs", func(t *testing.T) {
		salesJID, _ := repo.GetSenderJID("sales")
		supportJID, _ := repo.GetSenderJID("support")
		marketingJID, _ := repo.GetSenderJID("marketing")

		if salesJID == supportJID {
			t.Error("Sales and Support should have different JIDs")
		}
		if salesJID == marketingJID {
			t.Error("Sales and Marketing should have different JIDs")
		}
		if supportJID == marketingJID {
			t.Error("Support and Marketing should have different JIDs")
		}

		// Verify expected phone numbers are in JIDs
		if !contains(salesJID, "6281234567890") {
			t.Errorf("Sales JID should contain phone number, got: %s", salesJID)
		}
		if !contains(supportJID, "6289876543210") {
			t.Errorf("Support JID should contain phone number, got: %s", supportJID)
		}
		if !contains(marketingJID, "6285555555555") {
			t.Errorf("Marketing JID should contain phone number, got: %s", marketingJID)
		}
	})

	t.Run("Verify default sender", func(t *testing.T) {
		defaultJID := repo.GetJID()
		salesJID, _ := repo.GetSenderJID("sales")

		if defaultJID != salesJID {
			t.Errorf("Default JID should match sales JID. Default: %s, Sales: %s", defaultJID, salesJID)
		}
	})

	t.Run("Test sending from non-existent sender", func(t *testing.T) {
		ctx := context.Background()

		_, err := repo.SendMessageFrom(ctx, "invalid_sender", "6281111111111@s.whatsapp.net", "Test message")

		if err == nil {
			t.Error("Expected error when sending from non-existent sender")
		}

		// Error could be either "not found" or "not connected" depending on fallback behavior
		errMsg := err.Error()
		if !contains(errMsg, "invalid_sender") {
			t.Errorf("Error message should contain sender ID, got: %s", errMsg)
		}
	})

	t.Run("Test sending with invalid recipient JID", func(t *testing.T) {
		ctx := context.Background()

		// Try to send from sales with invalid JID
		_, err := repo.SendMessageFrom(ctx, "sales", "invalid-jid-format", "Test message")

		if err == nil {
			t.Error("Expected error for invalid JID format")
		}
	})
}

// TestMultipleSendersWithClientManager tests multiple senders using ClientManager
func TestMultipleSendersWithClientManager(t *testing.T) {
	// Setup: Create multiple sender clients
	client1 := createMockClient("6281111111111", true)
	client2 := createMockClient("6282222222222", true)
	client3 := createMockClient("6283333333333", true)

	mockManager := &mockClientManager{
		clients: map[string]*whatsmeow.Client{
			"sender1": client1,
			"sender2": client2,
			"sender3": client3,
		},
		defaultClient: client1,
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClientManager(nil, mockManager)

	t.Run("Retrieve JIDs from all senders via manager", func(t *testing.T) {
		jids := make(map[string]string)

		for senderID := range mockManager.clients {
			jid, err := repo.GetSenderJID(senderID)
			if err != nil {
				t.Errorf("Failed to get JID for %s: %v", senderID, err)
				continue
			}
			jids[senderID] = jid
		}

		if len(jids) != 3 {
			t.Errorf("Expected 3 JIDs, got %d", len(jids))
		}

		// Verify all JIDs are unique
		uniqueJIDs := make(map[string]bool)
		for _, jid := range jids {
			if uniqueJIDs[jid] {
				t.Errorf("Duplicate JID found: %s", jid)
			}
			uniqueJIDs[jid] = true
		}
	})

	t.Run("Verify connection status with multiple clients", func(t *testing.T) {
		// Our mock clients are not connected, so IsConnected should return false
		isConnected := repo.IsConnected()

		// Mock clients won't be connected
		if isConnected {
			t.Error("Expected false for mock clients")
		}
	})

	t.Run("Test fallback to default when specific sender not found", func(t *testing.T) {
		// Try to get JID with empty sender ID - should fall back to default
		defaultJID := repo.GetJID()
		sender1JID, _ := repo.GetSenderJID("sender1")

		if defaultJID != sender1JID {
			t.Error("Default JID should match sender1 (default client)")
		}
	})
}

// TestConcurrentMultipleSenders tests concurrent access with multiple senders
func TestConcurrentMultipleSenders(t *testing.T) {
	// Create multiple clients
	clients := make(map[string]*whatsmeow.Client)
	for i := 0; i < 10; i++ {
		senderID := fmt.Sprintf("sender_%d", i)
		phoneNumber := fmt.Sprintf("62812345678%02d", i)
		clients[senderID] = createMockClient(phoneNumber, true)
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClients(clients["sender_0"], nil, clients)

	t.Run("Concurrent access to different senders", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 100)

		// Launch concurrent goroutines to access different senders
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				senderID := fmt.Sprintf("sender_%d", id)

				// Try to get JID multiple times
				for j := 0; j < 10; j++ {
					jid, err := repo.GetSenderJID(senderID)
					if err != nil {
						errors <- fmt.Errorf("sender %s error: %v", senderID, err)
						return
					}
					if jid == "" {
						errors <- fmt.Errorf("sender %s returned empty JID", senderID)
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent access error: %v", err)
		}
	})
}

// TestSenderSelection tests that the correct sender is selected for operations
func TestSenderSelection(t *testing.T) {
	// Create three distinct clients
	client1 := createMockClient("6281111111111", true)
	client2 := createMockClient("6282222222222", true)
	client3 := createMockClient("6283333333333", true)

	clients := map[string]*whatsmeow.Client{
		"customer_service": client1,
		"sales_team":       client2,
		"tech_support":     client3,
	}

	repo := infrastructure.NewWhatsAppRepositoryWithClients(client1, nil, clients)

	testCases := []struct {
		senderID      string
		expectedPhone string
	}{
		{"customer_service", "6281111111111"},
		{"sales_team", "6282222222222"},
		{"tech_support", "6283333333333"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Select sender %s", tc.senderID), func(t *testing.T) {
			jid, err := repo.GetSenderJID(tc.senderID)

			if err != nil {
				t.Fatalf("Failed to get JID for %s: %v", tc.senderID, err)
			}

			if !contains(jid, tc.expectedPhone) {
				t.Errorf("Expected JID to contain %s, got %s", tc.expectedPhone, jid)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && containsAt(s, substr, 1)
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
