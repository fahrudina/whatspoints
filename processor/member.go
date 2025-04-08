package processor

import (
	"database/sql"
	"fmt"

	"github.com/wa-serv/repository"
)

// GetMemberIDByPhoneNumber retrieves the member ID for a given phone number
func GetMemberIDByPhoneNumber(db *sql.DB, phoneNumber string) (int, error) {
	// Extract the phone number (remove any suffix like "@s.whatsapp.net")
	extractedPhoneNumber := extractPhoneNumber(phoneNumber)

	memberID, err := repository.GetMemberIDByPhoneNumber(db, extractedPhoneNumber)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve member ID: %w", err)
	}
	return memberID, nil
}

// GetMemberDetailsByPhoneNumber retrieves the member ID and name for a given phone number
func GetMemberDetailsByPhoneNumber(db *sql.DB, phoneNumber string) (int, string, error) {
	// Extract the phone number (remove any suffix like "@s.whatsapp.net")
	extractedPhoneNumber := extractPhoneNumber(phoneNumber)

	memberID, memberName, err := repository.GetMemberDetailsByPhoneNumber(db, extractedPhoneNumber)
	if err != nil {
		return 0, "", fmt.Errorf("failed to retrieve member details: %w", err)
	}
	return memberID, memberName, nil
}
