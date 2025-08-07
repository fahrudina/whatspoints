package repository

import (
	"database/sql"
	"fmt"
	"time"
)

// Member represents a user in the MEMBER table
type Member struct {
	MemberID    int
	PhoneNumber string
	Name        string
	Address     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RegisterMember adds a new member to the database
func RegisterMember(db *sql.DB, name, address, phoneNumber string) error {
	// Start a transaction for member registration
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Insert into MEMBER table with current timestamp and return the member ID
	query := `INSERT INTO members (name, address, phone_number, created_at, updated_at) 
              VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING member_id`

	var memberID int
	err = tx.QueryRow(query, name, address, phoneNumber).Scan(&memberID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to register member: %v", err)
	}

	// Create initial point record for the member
	pointQuery := `INSERT INTO points (member_id, accumulated_points, current_points, created_at, updated_at) 
                   VALUES ($1, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err = tx.Exec(pointQuery, memberID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to initialize points: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// IsMemberRegistered checks if a user is already registered
func IsMemberRegistered(db *sql.DB, phoneNumber string) (bool, error) {
	query := `SELECT COUNT(*) FROM members WHERE phone_number = $1`

	var count int
	err := db.QueryRow(query, phoneNumber).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetMemberIDByPhoneNumber retrieves the member_id for a given phone number
func GetMemberIDByPhoneNumber(db *sql.DB, phoneNumber string) (int, error) {
	var memberID int
	query := "SELECT member_id FROM members WHERE phone_number = $1"
	err := db.QueryRow(query, phoneNumber).Scan(&memberID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no member found with phone number: %s", phoneNumber)
		}
		return 0, fmt.Errorf("failed to retrieve member ID: %w", err)
	}
	return memberID, nil
}

// GetMemberNameByID retrieves the member's name for a given member ID
func GetMemberNameByID(db *sql.DB, memberID int) (string, error) {
	var memberName string
	query := "SELECT name FROM members WHERE member_id = $1"
	err := db.QueryRow(query, memberID).Scan(&memberName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no member found with ID: %d", memberID)
		}
		return "", fmt.Errorf("failed to retrieve member name: %w", err)
	}
	return memberName, nil
}

// GetMemberDetailsByPhoneNumber retrieves the member ID and name for a given phone number
func GetMemberDetailsByPhoneNumber(db *sql.DB, phoneNumber string) (int, string, error) {
	var memberID int
	var memberName string
	query := "SELECT member_id, name FROM members WHERE phone_number = $1"
	err := db.QueryRow(query, phoneNumber).Scan(&memberID, &memberName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", fmt.Errorf("no member found with phone number: %s", phoneNumber)
		}
		return 0, "", fmt.Errorf("failed to retrieve member details: %w", err)
	}
	return memberID, memberName, nil
}
