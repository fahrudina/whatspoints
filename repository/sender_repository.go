package repository

import (
	"database/sql"
	"fmt"
	"time"
)

// Sender represents a WhatsApp sender in the database
type Sender struct {
	SenderID    string
	PhoneNumber string
	Name        string
	IsDefault   bool
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateSenderIfNotExists creates a sender record if it doesn't already exist
func CreateSenderIfNotExists(db *sql.DB, senderID, phoneNumber, name string, isDefault bool) error {
	query := `
		INSERT INTO senders (sender_id, phone_number, name, is_default, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (sender_id) DO NOTHING
	`

	_, err := db.Exec(query, senderID, phoneNumber, name, isDefault, true)
	if err != nil {
		return fmt.Errorf("failed to create sender record: %w", err)
	}

	return nil
}

// GetSenderByID retrieves a sender by their ID
func GetSenderByID(db *sql.DB, senderID string) (*Sender, error) {
	query := `
		SELECT sender_id, phone_number, name, is_default, is_active, created_at, updated_at
		FROM senders
		WHERE sender_id = $1
	`

	var sender Sender
	err := db.QueryRow(query, senderID).Scan(
		&sender.SenderID,
		&sender.PhoneNumber,
		&sender.Name,
		&sender.IsDefault,
		&sender.IsActive,
		&sender.CreatedAt,
		&sender.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("sender not found: %s", senderID)
		}
		return nil, fmt.Errorf("failed to get sender: %w", err)
	}

	return &sender, nil
}

// GetDefaultSender retrieves the default sender from the database
func GetDefaultSender(db *sql.DB) (*Sender, error) {
	query := `
		SELECT sender_id, phone_number, name, is_default, is_active, created_at, updated_at
		FROM senders
		WHERE is_default = true AND is_active = true
		LIMIT 1
	`

	var sender Sender
	err := db.QueryRow(query).Scan(
		&sender.SenderID,
		&sender.PhoneNumber,
		&sender.Name,
		&sender.IsDefault,
		&sender.IsActive,
		&sender.CreatedAt,
		&sender.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// If no default sender found, try to get the first active sender
			return getFirstActiveSender(db)
		}
		return nil, fmt.Errorf("failed to get default sender: %w", err)
	}

	return &sender, nil
}

// getFirstActiveSender retrieves the first active sender ordered by creation date
func getFirstActiveSender(db *sql.DB) (*Sender, error) {
	query := `
		SELECT sender_id, phone_number, name, is_default, is_active, created_at, updated_at
		FROM senders
		WHERE is_active = true
		ORDER BY created_at ASC
		LIMIT 1
	`

	var sender Sender
	err := db.QueryRow(query).Scan(
		&sender.SenderID,
		&sender.PhoneNumber,
		&sender.Name,
		&sender.IsDefault,
		&sender.IsActive,
		&sender.CreatedAt,
		&sender.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active senders found")
		}
		return nil, fmt.Errorf("failed to get first active sender: %w", err)
	}

	return &sender, nil
}

// GetAllSenders retrieves all senders from the database
func GetAllSenders(db *sql.DB) ([]Sender, error) {
	query := `
		SELECT sender_id, phone_number, name, is_default, is_active, created_at, updated_at
		FROM senders
		ORDER BY is_default DESC, created_at ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query senders: %w", err)
	}
	defer rows.Close()

	var senders []Sender
	for rows.Next() {
		var sender Sender
		err := rows.Scan(
			&sender.SenderID,
			&sender.PhoneNumber,
			&sender.Name,
			&sender.IsDefault,
			&sender.IsActive,
			&sender.CreatedAt,
			&sender.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sender: %w", err)
		}
		senders = append(senders, sender)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating senders: %w", err)
	}

	return senders, nil
}

// UpdateSenderStatus updates the active status of a sender
func UpdateSenderStatus(db *sql.DB, senderID string, isActive bool) error {
	query := `
		UPDATE senders
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE sender_id = $2
	`

	result, err := db.Exec(query, isActive, senderID)
	if err != nil {
		return fmt.Errorf("failed to update sender status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("sender not found: %s", senderID)
	}

	return nil
}

// SetDefaultSender sets a sender as the default sender and unsets all others
func SetDefaultSender(db *sql.DB, senderID string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Unset all default flags
	_, err = tx.Exec("UPDATE senders SET is_default = false, updated_at = CURRENT_TIMESTAMP")
	if err != nil {
		return fmt.Errorf("failed to unset default flags: %w", err)
	}

	// Set the new default
	result, err := tx.Exec(
		"UPDATE senders SET is_default = true, updated_at = CURRENT_TIMESTAMP WHERE sender_id = $1",
		senderID,
	)
	if err != nil {
		return fmt.Errorf("failed to set default sender: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("sender not found: %s", senderID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
