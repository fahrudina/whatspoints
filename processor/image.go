package processor

import (
	"database/sql"
	"fmt"

	"github.com/wa-serv/repository"
)

// SaveImageURL saves the image URL for a member
func SaveImageURL(db *sql.DB, memberID int, imageURL string) error {
	err := repository.SaveImageURL(db, memberID, imageURL)
	if err != nil {
		return fmt.Errorf("failed to save image URL: %w", err)
	}
	return nil
}
