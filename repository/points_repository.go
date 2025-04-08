package repository

import (
	"fmt"
)

// Executor interface to abstract the database operations
// type Executor interface {
// 	QueryRow(query string, args ...interface{}) *sql.Row
// 	Exec(query string, args ...interface{}) (sql.Result, error)
// }

// GetCurrentPoints retrieves the current points for a member by their ID
func GetCurrentPoints(exec Executor, memberID int) (int, error) {
	var currentPoints int
	query := "SELECT current_points FROM points WHERE member_id = ?"
	err := exec.QueryRow(query, memberID).Scan(&currentPoints)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0, fmt.Errorf("no points record found for member ID: %d", memberID)
		}
		return 0, fmt.Errorf("failed to retrieve current points: %w", err)
	}
	return currentPoints, nil
}

// UpsertPoints performs an upsert operation for the points table
func UpsertPoints(exec Executor, memberID, currentPoints int) error {
	query := `
	INSERT INTO points (member_id, accumulated_points, current_points)
	VALUES (?, ?, ?)
	ON DUPLICATE KEY UPDATE
		accumulated_points = accumulated_points + VALUES(current_points),
		current_points = current_points + VALUES(current_points)
	`
	_, err := exec.Exec(query, memberID, currentPoints, currentPoints)
	if err != nil {
		return fmt.Errorf("failed to upsert points: %w", err)
	}
	return nil
}

// DeductPoints deducts points from the current_points column
func DeductPoints(exec Executor, memberID, pointsToDeduct int) error {
	query := `
	UPDATE points
	SET current_points = current_points - ?
	WHERE member_id = ?
	`
	_, err := exec.Exec(query, pointsToDeduct, memberID)
	if err != nil {
		return fmt.Errorf("failed to deduct points: %w", err)
	}
	return nil
}
