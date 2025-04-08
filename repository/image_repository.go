package repository

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
