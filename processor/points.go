package processor

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/wa-serv/config"
	"github.com/wa-serv/repository"
)

// ProcessUpsertPoints handles the upsert points action
func ProcessUpsertPoints(db *sql.DB, senderPhoneNumber, input string) error {
	senderPhoneNumber = extractPhoneNumber(senderPhoneNumber)
	// Check if the sender is allowed to perform this action
	if !config.Env.AllowedPhoneNumbers[senderPhoneNumber] {
		return errors.New("unauthorized action: phone number not allowed")
	}

	// Parse the input
	parts := strings.Split(input, "#")
	if len(parts) != 3 {
		return errors.New("invalid input format: expected INPUT#phone_number#current_points")
	}

	phoneNumber := parts[1]
	currentPoints, err := parsePoints(parts[2])
	if err != nil {
		return fmt.Errorf("invalid points value: %w", err)
	}

	// Get the member ID by phone number
	memberID, err := GetMemberIDByPhoneNumber(db, phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to retrieve member ID: %w", err)
	}

	// Upsert points for the member and track the transaction
	err = upsertPointsWithTransaction(db, memberID, currentPoints)
	if err != nil {
		return fmt.Errorf("failed to upsert points: %w", err)
	}

	return nil
}

// parsePoints parses the points value from a string
func parsePoints(pointsStr string) (int, error) {
	var points int
	_, err := fmt.Sscanf(pointsStr, "%d", &points)
	if err != nil {
		return 0, err
	}
	return points, nil
}

// upsertPointsWithTransaction performs an upsert operation for the points table and tracks the transaction
func upsertPointsWithTransaction(db *sql.DB, memberID, currentPoints int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Upsert points
	err = repository.UpsertPoints(tx, memberID, currentPoints)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Track the transaction in point_transactions
	err = repository.InsertPointTransaction(tx, memberID, currentPoints, "EARN", "Points updated via upsert")
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetCurrentPoints retrieves the current points for a member by their ID
func GetCurrentPoints(db *sql.DB, memberID int) (int, error) {
	var currentPoints int
	query := "SELECT current_points FROM points WHERE member_id = $1"
	err := db.QueryRow(query, memberID).Scan(&currentPoints)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no points record found for member ID: %d", memberID)
		}
		return 0, fmt.Errorf("failed to retrieve current points: %w", err)
	}
	return currentPoints, nil
}
