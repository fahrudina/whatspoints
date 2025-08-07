-- Whatsmeow Session Storage Tables Migration for PostgreSQL/Supabase
-- Run this script to create the required tables for WhatsApp session storage

-- Device table for storing device information including push names
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
);

-- Identity keys table
CREATE TABLE IF NOT EXISTS whatsmeow_identity_keys (
    our_jid TEXT,
    their_id TEXT,
    identity BYTEA,
    PRIMARY KEY (our_jid, their_id)
);

-- Pre-keys table
CREATE TABLE IF NOT EXISTS whatsmeow_pre_keys (
    jid TEXT,
    key_id INTEGER,
    key BYTEA,
    uploaded BOOLEAN,
    PRIMARY KEY (jid, key_id)
);

-- Sessions table
CREATE TABLE IF NOT EXISTS whatsmeow_sessions (
    our_jid TEXT,
    their_id TEXT,
    session BYTEA,
    PRIMARY KEY (our_jid, their_id)
);

-- Sender keys table
CREATE TABLE IF NOT EXISTS whatsmeow_sender_keys (
    our_jid TEXT,
    chat_id TEXT,
    sender_id TEXT,
    sender_key BYTEA,
    PRIMARY KEY (our_jid, chat_id, sender_id)
);

-- App state sync keys table
CREATE TABLE IF NOT EXISTS whatsmeow_app_state_sync_keys (
    jid TEXT,
    key_id BYTEA,
    key_data BYTEA,
    timestamp BIGINT,
    fingerprint BYTEA,
    PRIMARY KEY (jid, key_id)
);

-- App state version table
CREATE TABLE IF NOT EXISTS whatsmeow_app_state_version (
    jid TEXT,
    name TEXT,
    version BIGINT,
    hash BYTEA,
    PRIMARY KEY (jid, name)
);

-- App state mutation macs table
CREATE TABLE IF NOT EXISTS whatsmeow_app_state_mutation_macs (
    jid TEXT,
    name TEXT,
    version BIGINT,
    index_mac BYTEA,
    value_mac BYTEA,
    PRIMARY KEY (jid, name, version, index_mac)
);

-- Contacts table
CREATE TABLE IF NOT EXISTS whatsmeow_contacts (
    our_jid TEXT,
    their_jid TEXT,
    first_name TEXT,
    full_name TEXT,
    push_name TEXT,
    business_name TEXT,
    PRIMARY KEY (our_jid, their_jid)
);

-- Chat settings table
CREATE TABLE IF NOT EXISTS whatsmeow_chat_settings (
    our_jid TEXT,
    chat_jid TEXT,
    muted_until BIGINT,
    pinned BOOLEAN,
    archived BOOLEAN,
    PRIMARY KEY (our_jid, chat_jid)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_whatsmeow_device_jid ON whatsmeow_device(jid);
CREATE INDEX IF NOT EXISTS idx_whatsmeow_contacts_our_jid ON whatsmeow_contacts(our_jid);
CREATE INDEX IF NOT EXISTS idx_whatsmeow_sessions_our_jid ON whatsmeow_sessions(our_jid);
CREATE INDEX IF NOT EXISTS idx_whatsmeow_pre_keys_jid ON whatsmeow_pre_keys(jid);

-- Success message
SELECT 'Whatsmeow tables created successfully!' as status;
