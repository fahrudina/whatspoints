package repository

import "database/sql"

// Executor defines the common interface for *sql.DB and *sql.Tx
type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}
