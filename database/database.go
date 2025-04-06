package database

import (
	"database/sql"
	"fmt"
)

// SaveImageURL saves the image URL to the database
func SaveImageURL(db *sql.DB, memberID int, imageURL string) error {
	query := "INSERT INTO images (member_id, image_url) VALUES (?, ?)"
	_, err := db.Exec(query, memberID, imageURL)
	if err != nil {
		return fmt.Errorf("failed to save image URL: %w", err)
	}
	return nil
}

// GetMemberIDByPhoneNumber retrieves the member_id for a given phone number
func GetMemberIDByPhoneNumber(db *sql.DB, phoneNumber string) (int, error) {
	var memberID int
	query := "SELECT member_id FROM members WHERE phone_number = ?"
	err := db.QueryRow(query, phoneNumber).Scan(&memberID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no member found with phone number: %s", phoneNumber)
		}
		return 0, fmt.Errorf("failed to retrieve member ID: %w", err)
	}
	return memberID, nil
}

// InitImageTable initializes the images table
func InitImageTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS images (
		image_id INT AUTO_INCREMENT PRIMARY KEY,
		member_id INT,
		image_url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (member_id) REFERENCES members(member_id)
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create images table: %w", err)
	}
	return nil
}

// InitMemberTable initializes the members table
func InitMemberTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS members (
		member_id INT AUTO_INCREMENT PRIMARY KEY,
		phone_number VARCHAR(20) UNIQUE,
		name VARCHAR(100),
		address TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create members table: %w", err)
	}
	return nil
}

// InitReceiptsTable initializes the receipts table
func InitReceiptsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS receipts (
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
		FOREIGN KEY (member_id) REFERENCES members(member_id)
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create receipts table: %w", err)
	}
	return nil
}

// InitPointsTable initializes the points table
func InitPointsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS points (
		point_id INT AUTO_INCREMENT PRIMARY KEY,
		member_id INT,
		accumulated_points INT,
		current_points INT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (member_id) REFERENCES members(member_id)
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create points table: %w", err)
	}
	return nil
}

// InitPointTransactionsTable initializes the point_transactions table
func InitPointTransactionsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS point_transactions (
		transaction_id INT AUTO_INCREMENT PRIMARY KEY,
		point_id INT,
		receipt_id INT NULL,
		points_changed INT,
		transaction_type VARCHAR(20),
		transaction_date DATETIME,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (point_id) REFERENCES points(point_id),
		FOREIGN KEY (receipt_id) REFERENCES receipts(receipt_id)
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create point_transactions table: %w", err)
	}
	return nil
}
