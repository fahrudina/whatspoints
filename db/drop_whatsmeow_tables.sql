-- Drop all WhatsApp session storage tables
-- Run this script to clean up existing whatsmeow tables before letting the library recreate them

DROP TABLE IF EXISTS whatsmeow_app_state_sync_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_app_state_version CASCADE;
DROP TABLE IF EXISTS whatsmeow_contacts CASCADE;
DROP TABLE IF EXISTS whatsmeow_message_secrets CASCADE;
DROP TABLE IF EXISTS whatsmeow_sender_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_sessions CASCADE;
DROP TABLE IF EXISTS whatsmeow_pre_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_identity_keys CASCADE;
DROP TABLE IF EXISTS whatsmeow_device CASCADE;

-- Note: After running this script, the whatsmeow library will automatically
-- recreate these tables with the correct schema when you start the application
