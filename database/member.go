package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
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

	// Insert into MEMBER table with current timestamp
	query := `INSERT INTO MEMBER (name, address, phone_number, created_at, updated_at) 
              VALUES (?, ?, ?, NOW(), NOW())`

	result, err := tx.Exec(query, name, address, phoneNumber)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to register member: %v", err)
	}

	// Get the inserted member ID
	memberID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get last insert ID: %v", err)
	}

	// Create initial point record for the member
	pointQuery := `INSERT INTO POINT (member_id, accumulated_points, current_points, created_at, updated_at) 
                   VALUES (?, 0, 0, NOW(), NOW())`
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
	query := `SELECT COUNT(*) FROM MEMBER WHERE phone_number = ?`

	var count int
	err := db.QueryRow(query, phoneNumber).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
