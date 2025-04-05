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

// InitMemberTable ensures all required tables exist in the database
// Note: This doesn't recreate tables if they already exist
func InitMemberTable(db *sql.DB) error {
	// Check if MEMBER table exists
	var tableExists int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'MEMBER'").Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check if MEMBER table exists: %v", err)
	}

	// If tables don't exist, create them following the init_table.sql schema
	if tableExists == 0 {
		// Create tables following the provided schema
		memberQuery := `CREATE TABLE MEMBER (
			member_id INT AUTO_INCREMENT PRIMARY KEY,
			phone_number VARCHAR(20) UNIQUE,
			name VARCHAR(100),
			address TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`

		_, err := db.Exec(memberQuery)
		if err != nil {
			return fmt.Errorf("failed to create MEMBER table: %v", err)
		}

		pointQuery := `CREATE TABLE POINT (
			point_id INT AUTO_INCREMENT PRIMARY KEY,
			member_id INT,
			accumulated_points INT DEFAULT 0,
			current_points INT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (member_id) REFERENCES MEMBER(member_id)
		)`

		_, err = db.Exec(pointQuery)
		if err != nil {
			return fmt.Errorf("failed to create POINT table: %v", err)
		}

		receiptQuery := `CREATE TABLE RECEIPT (
			receipt_id INT AUTO_INCREMENT PRIMARY KEY,
			member_id INT,
			receipt_image TEXT,
			total_kg DECIMAL(10, 2),
			total_unit INT,
			total_price DECIMAL(10, 2),
			points_earned INT,
			receipt_date DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (member_id) REFERENCES MEMBER(member_id)
		)`

		_, err = db.Exec(receiptQuery)
		if err != nil {
			return fmt.Errorf("failed to create RECEIPT table: %v", err)
		}

		transactionQuery := `CREATE TABLE POINT_TRANSACTION (
			transaction_id INT AUTO_INCREMENT PRIMARY KEY,
			point_id INT,
			receipt_id INT NULL,
			points_changed INT,
			transaction_type VARCHAR(20),
			transaction_date DATETIME DEFAULT CURRENT_TIMESTAMP,
			notes TEXT,
			FOREIGN KEY (point_id) REFERENCES POINT(point_id),
			FOREIGN KEY (receipt_id) REFERENCES RECEIPT(receipt_id)
		)`

		_, err = db.Exec(transactionQuery)
		if err != nil {
			return fmt.Errorf("failed to create POINT_TRANSACTION table: %v", err)
		}

		fmt.Println("All database tables created successfully")
	}

	return nil
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
