package repository

import (
	"fmt"
)

// InsertPointTransaction logs a transaction in the point_transactions table
func InsertPointTransaction(exec Executor, memberID, pointsChanged int, transactionType, notes string) error {
	query := `
	INSERT INTO point_transactions (point_id, points_changed, transaction_type, transaction_date, notes)
	VALUES (
		(SELECT point_id FROM points WHERE member_id = $1),
		$2, $3, CURRENT_TIMESTAMP, $4
	)
	`
	_, err := exec.Exec(query, memberID, pointsChanged, transactionType, notes)
	if err != nil {
		return fmt.Errorf("failed to insert point transaction: %w", err)
	}
	return nil
}
