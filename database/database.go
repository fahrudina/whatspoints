package database

import (
	"database/sql"
	"fmt"
)

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
		member_id INT UNIQUE, -- Set member_id as unique
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

// InitItemsTable initializes the items table
func InitItemsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS items (
		item_id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		price_per_unit DECIMAL(10, 2) DEFAULT 0.00,
		price_per_kilo DECIMAL(10, 2) DEFAULT 0.00,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create items table: %w", err)
	}
	return nil
}

// InitOrderItemsTable initializes the order_items table
func InitOrderItemsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS order_items (
		order_item_id INT AUTO_INCREMENT PRIMARY KEY,
		order_id INT,
		item_id INT,
		total_kilo DECIMAL(10, 2) DEFAULT 0.00,
		total_unit INT DEFAULT 0,
		price DECIMAL(10, 2),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (order_id) REFERENCES orders(order_id),
		FOREIGN KEY (item_id) REFERENCES items(item_id)
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create order_items table: %w", err)
	}
	return nil
}

// InitOrdersTable initializes the orders table
func InitOrdersTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS orders (
		order_id INT AUTO_INCREMENT PRIMARY KEY,
		member_id INT,
		total_price DECIMAL(10, 2),
		order_date DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (member_id) REFERENCES members(member_id)
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}
	return nil
}
