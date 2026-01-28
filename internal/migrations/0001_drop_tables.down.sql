-- migrations/0001_drop_tables.down.sql

DROP TABLE IF EXISTS otp_codes;
DROP TABLE IF EXISTS users;

DROP INDEX IF EXISTS idx_users_telegram_id;
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_otp_codes_phone;
DROP INDEX IF EXISTS idx_otp_codes_code;
DROP INDEX IF EXISTS idx_otp_codes_expires_at;