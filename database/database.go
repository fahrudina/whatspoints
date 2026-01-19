package database

import (
	"database/sql"
	"fmt"
	"os"
)

// BuildPostgresConnectionString builds a PostgreSQL connection string from environment variables
func BuildPostgresConnectionString() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s&statement_cache_mode=describe&default_query_exec_mode=simple_protocol",
		os.Getenv("SUPABASE_USER"),
		os.Getenv("SUPABASE_PASSWORD"),
		os.Getenv("SUPABASE_HOST"),
		os.Getenv("SUPABASE_PORT"),
		os.Getenv("SUPABASE_DB"),
		getSSLMode(),
	)
}

// getSSLMode returns the SSL mode from environment variable or default
func getSSLMode() string {
	sslMode := os.Getenv("SUPABASE_SSLMODE")
	if sslMode == "" {
		return "require"
	}
	return sslMode
}

// InitImageTable initializes the images table
func InitImageTable(db *sql.DB) error {
	query := `
	   CREATE TABLE IF NOT EXISTS images (
			   image_id SERIAL PRIMARY KEY,
			   member_id INTEGER,
			   image_url TEXT NOT NULL,
			   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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
			   member_id SERIAL PRIMARY KEY,
			   phone_number VARCHAR(20) UNIQUE,
			   name VARCHAR(100),
			   address TEXT,
			   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
			   receipt_id SERIAL PRIMARY KEY,
			   member_id INTEGER,
			   receipt_image TEXT,
			   total_kg NUMERIC(10, 2),
			   total_unit INTEGER,
			   total_price NUMERIC(10, 2),
			   points_earned INTEGER,
			   receipt_date TIMESTAMP,
			   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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
			   point_id SERIAL PRIMARY KEY,
			   member_id INTEGER UNIQUE,
			   accumulated_points INTEGER,
			   current_points INTEGER,
			   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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
			   transaction_id SERIAL PRIMARY KEY,
			   point_id INTEGER,
			   receipt_id INTEGER,
			   points_changed INTEGER,
			   transaction_type VARCHAR(20),
			   transaction_date TIMESTAMP,
			   notes TEXT,
			   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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
			   item_id SERIAL PRIMARY KEY,
			   name VARCHAR(100) NOT NULL,
			   description TEXT,
			   price_per_unit NUMERIC(10, 2) DEFAULT 0.00,
			   price_per_kilo NUMERIC(10, 2) DEFAULT 0.00,
			   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
			   order_item_id SERIAL PRIMARY KEY,
			   order_id INTEGER,
			   item_id INTEGER,
			   total_kilo NUMERIC(10, 2) DEFAULT 0.00,
			   total_unit INTEGER DEFAULT 0,
			   price NUMERIC(10, 2),
			   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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
			   order_id SERIAL PRIMARY KEY,
			   member_id INTEGER,
			   total_price NUMERIC(10, 2),
			   order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			   FOREIGN KEY (member_id) REFERENCES members(member_id)
	   )`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}
	return nil
}

// InitWhatsmeowTables initializes the required tables for Whatsmeow session storage
func InitWhatsmeowTables(db *sql.DB) error {
	// Create the whatsmeow_device table
	deviceTableQuery := `
	CREATE TABLE IF NOT EXISTS whatsmeow_device (
		jid TEXT PRIMARY KEY,
		registration_id INTEGER NOT NULL,
		noise_key BYTEA NOT NULL,
		identity_key BYTEA NOT NULL,
		signed_pre_key BYTEA NOT NULL,
		signed_pre_key_id INTEGER NOT NULL,
		signed_pre_key_sig BYTEA NOT NULL,
		adv_key BYTEA,
		adv_details BYTEA,
		adv_account_sig BYTEA,
		adv_account_sig_key BYTEA,
		adv_device_sig BYTEA,
		platform TEXT NOT NULL DEFAULT '',
		business_name TEXT NOT NULL DEFAULT '',
		push_name TEXT NOT NULL DEFAULT ''
	)`

	if _, err := db.Exec(deviceTableQuery); err != nil {
		return fmt.Errorf("failed to create whatsmeow_device table: %w", err)
	}

	// Create the whatsmeow_identity_keys table
	identityTableQuery := `
	CREATE TABLE IF NOT EXISTS whatsmeow_identity_keys (
		our_jid TEXT,
		their_id TEXT,
		identity BYTEA,
		PRIMARY KEY (our_jid, their_id)
	)`

	if _, err := db.Exec(identityTableQuery); err != nil {
		return fmt.Errorf("failed to create whatsmeow_identity_keys table: %w", err)
	}

	// Create the whatsmeow_pre_keys table
	preKeysTableQuery := `
	CREATE TABLE IF NOT EXISTS whatsmeow_pre_keys (
		jid TEXT,
		key_id INTEGER,
		key BYTEA,
		uploaded BOOLEAN,
		PRIMARY KEY (jid, key_id)
	)`

	if _, err := db.Exec(preKeysTableQuery); err != nil {
		return fmt.Errorf("failed to create whatsmeow_pre_keys table: %w", err)
	}

	// Create the whatsmeow_sessions table
	sessionsTableQuery := `
	CREATE TABLE IF NOT EXISTS whatsmeow_sessions (
		our_jid TEXT,
		their_id TEXT,
		session BYTEA,
		PRIMARY KEY (our_jid, their_id)
	)`

	if _, err := db.Exec(sessionsTableQuery); err != nil {
		return fmt.Errorf("failed to create whatsmeow_sessions table: %w", err)
	}

	// Create the whatsmeow_sender_keys table
	senderKeysTableQuery := `
	CREATE TABLE IF NOT EXISTS whatsmeow_sender_keys (
		our_jid TEXT,
		chat_id TEXT,
		sender_id TEXT,
		sender_key BYTEA,
		PRIMARY KEY (our_jid, chat_id, sender_id)
	)`

	if _, err := db.Exec(senderKeysTableQuery); err != nil {
		return fmt.Errorf("failed to create whatsmeow_sender_keys table: %w", err)
	}

	// Create the whatsmeow_app_state_sync_keys table
	appStateSyncKeysTableQuery := `
	CREATE TABLE IF NOT EXISTS whatsmeow_app_state_sync_keys (
		jid TEXT,
		key_id BYTEA,
		key_data BYTEA,
		timestamp BIGINT,
		fingerprint BYTEA,
		PRIMARY KEY (jid, key_id)
	)`

	if _, err := db.Exec(appStateSyncKeysTableQuery); err != nil {
		return fmt.Errorf("failed to create whatsmeow_app_state_sync_keys table: %w", err)
	}

	// Create the whatsmeow_app_state_version table
	appStateVersionTableQuery := `
	CREATE TABLE IF NOT EXISTS whatsmeow_app_state_version (
		jid TEXT,
		name TEXT,
		version BIGINT,
		hash BYTEA,
		PRIMARY KEY (jid, name)
	)`

	if _, err := db.Exec(appStateVersionTableQuery); err != nil {
		return fmt.Errorf("failed to create whatsmeow_app_state_version table: %w", err)
	}

	// Create the whatsmeow_app_state_mutation_macs table
	appStateMutationMacsTableQuery := `
	CREATE TABLE IF NOT EXISTS whatsmeow_app_state_mutation_macs (
		jid TEXT,
		name TEXT,
		version BIGINT,
		index_mac BYTEA,
		value_mac BYTEA,
		PRIMARY KEY (jid, name, version, index_mac)
	)`

	if _, err := db.Exec(appStateMutationMacsTableQuery); err != nil {
		return fmt.Errorf("failed to create whatsmeow_app_state_mutation_macs table: %w", err)
	}

	// Create the whatsmeow_contacts table
	contactsTableQuery := `
	CREATE TABLE IF NOT EXISTS whatsmeow_contacts (
		our_jid TEXT,
		their_jid TEXT,
		first_name TEXT,
		full_name TEXT,
		push_name TEXT,
		business_name TEXT,
		PRIMARY KEY (our_jid, their_jid)
	)`

	if _, err := db.Exec(contactsTableQuery); err != nil {
		return fmt.Errorf("failed to create whatsmeow_contacts table: %w", err)
	}

	// Create the whatsmeow_chat_settings table
	chatSettingsTableQuery := `
	CREATE TABLE IF NOT EXISTS whatsmeow_chat_settings (
		our_jid TEXT,
		chat_jid TEXT,
		muted_until BIGINT,
		pinned BOOLEAN,
		archived BOOLEAN,
		PRIMARY KEY (our_jid, chat_jid)
	)`

	if _, err := db.Exec(chatSettingsTableQuery); err != nil {
		return fmt.Errorf("failed to create whatsmeow_chat_settings table: %w", err)
	}

	return nil
}

// InitSendersTable initializes the senders table for managing multiple WhatsApp sender accounts
func InitSendersTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS senders (
		sender_id VARCHAR(50) PRIMARY KEY,
		phone_number VARCHAR(30) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		is_default BOOLEAN DEFAULT FALSE,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create senders table: %w", err)
	}
	return nil
}
