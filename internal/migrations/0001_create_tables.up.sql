-- migrations/0001_create_tables.up.sql

-- Users table
CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     telegram_id BIGINT UNIQUE NOT NULL,
                                     username VARCHAR(255),
    language_code VARCHAR(10) DEFAULT 'uz',
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    region VARCHAR(255),
    district VARCHAR(255),
    school VARCHAR(255),
    grade INTEGER,
    phone VARCHAR(20),
    is_verified BOOLEAN DEFAULT FALSE,
    state VARCHAR(50) DEFAULT 'start',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

-- OTP codes table
CREATE TABLE IF NOT EXISTS otp_codes (
                                         id SERIAL PRIMARY KEY,
                                         user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    phone VARCHAR(20) NOT NULL,
    code VARCHAR(6) NOT NULL,
    message_id VARCHAR(100),
    is_used BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

-- Admin users table
CREATE TABLE IF NOT EXISTS admins (
                                      id SERIAL PRIMARY KEY,
                                      telegram_id BIGINT UNIQUE NOT NULL,
                                      username VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

-- Mandatory channels table
CREATE TABLE IF NOT EXISTS channels (
                                        id SERIAL PRIMARY KEY,
                                        channel_id BIGINT,
                                        channel_username VARCHAR(255) NOT NULL,
    title VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_otp_codes_phone ON otp_codes(phone);
CREATE INDEX IF NOT EXISTS idx_otp_codes_code ON otp_codes(code);
CREATE INDEX IF NOT EXISTS idx_admins_telegram_id ON admins(telegram_id);
CREATE INDEX IF NOT EXISTS idx_channels_active ON channels(is_active);

-- Insert default admin (o'zgartiring o'zingizning telegram_id ga)
INSERT INTO admins (telegram_id, username) VALUES (8485928313, 'BaxtiyorUrolov') ON CONFLICT DO NOTHING;