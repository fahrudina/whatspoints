package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Initialize tables
	if err := InitMemberTable(db); err != nil {
		return nil, err
	}
	if err := InitImageTable(db); err != nil {
		return nil, err
	}
	return db, nil
}

func TestSaveImageURL(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer db.Close()

	// Insert a test member
	_, err = db.Exec("INSERT INTO members (phone_number, name, address) VALUES (?, ?, ?)", "1234567890", "Test User", "Test Address")
	if err != nil {
		t.Fatalf("Failed to insert test member: %v", err)
	}

	// Get the member ID
	var memberID int
	err = db.QueryRow("SELECT member_id FROM members WHERE phone_number = ?", "1234567890").Scan(&memberID)
	if err != nil {
		t.Fatalf("Failed to retrieve member ID: %v", err)
	}

	// Test saving an image URL
	err = SaveImageURL(db, memberID, "http://example.com/image.jpg")
	if err != nil {
		t.Errorf("Failed to save image URL: %v", err)
	}

	// Verify the image URL was saved
	var imageURL string
	err = db.QueryRow("SELECT image_url FROM images WHERE member_id = ?", memberID).Scan(&imageURL)
	if err != nil {
		t.Errorf("Failed to retrieve image URL: %v", err)
	}
	if imageURL != "http://example.com/image.jpg" {
		t.Errorf("Expected image URL to be 'http://example.com/image.jpg', got '%s'", imageURL)
	}
}

func TestGetMemberIDByPhoneNumber(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer db.Close()

	// Insert a test member
	_, err = db.Exec("INSERT INTO members (phone_number, name, address) VALUES (?, ?, ?)", "1234567890", "Test User", "Test Address")
	if err != nil {
		t.Fatalf("Failed to insert test member: %v", err)
	}

	// Test retrieving the member ID
	memberID, err := GetMemberIDByPhoneNumber(db, "1234567890")
	if err != nil {
		t.Errorf("Failed to retrieve member ID: %v", err)
	}
	if memberID == 0 {
		t.Errorf("Expected a valid member ID, got 0")
	}
}
