package processor

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/wa-serv/repository"
)

var (
	ErrInsufficientPoints = errors.New("insufficient points for redemption")
	ErrMinimumPoints      = errors.New("minimum points required for redemption is 20")
	ErrInvalidPoints      = errors.New("invalid points value for redemption")
)

// RewardMapping defines the rewards for specific point values
var RewardMapping = map[int]string{
	20:  "Gratis cuci 2 kg",
	50:  "Gratis cuci 5 kg",
	100: "Pewangi premium atau gratis cuci 10 kg",
	150: "Voucher belanja Rp75.000",
	200: "Uang tunai Rp100.000 (dapat ditransfer ke rekening atau e-wallet)",
}

// RedeemPoints handles the redemption of points for a member and returns the reward
func RedeemPoints(db *sql.DB, phoneNumber string, pointsToRedeem int) (string, error) {
	// Enforce minimum points rule
	if pointsToRedeem < 20 {
		return "", ErrMinimumPoints
	}

	// Check if the points to redeem match a valid reward
	reward, exists := RewardMapping[pointsToRedeem]
	if !exists {
		return "", ErrInvalidPoints
	}

	// Get the member ID by phone number
	memberID, err := GetMemberIDByPhoneNumber(db, phoneNumber)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve member ID: %w", err)
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Check if the member has enough points
	currentPoints, err := repository.GetCurrentPoints(tx, memberID)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	if currentPoints < pointsToRedeem {
		tx.Rollback()
		return "", ErrInsufficientPoints
	}

	// Deduct the points
	err = repository.DeductPoints(tx, memberID, pointsToRedeem)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	// Track the redemption in point_transactions
	err = repository.InsertPointTransaction(tx, memberID, -pointsToRedeem, "REDEEM", fmt.Sprintf("Redeemed for: %s", reward))
	if err != nil {
		tx.Rollback()
		return "", err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	return reward, nil
}
