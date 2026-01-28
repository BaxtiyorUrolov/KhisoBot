// config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Environment string

	// Telegram
	TelegramBotToken string

	// Database
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string

	// SMS Service (sms.etc.uz)
	SMSBaseURL  string
	SMSLogin    string
	SMSPassword string
	SMSSender   string // CgPN - Amity

	// OTP Settings
	OTPLength      int
	OTPExpiresMins int
}

func Load() (*Config, error) {
	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),

		// Telegram
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),

		// Database
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:       getEnv("POSTGRES_DB", "khisobot"),
		PostgresSSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),

		// SMS Service
		SMSBaseURL:  getEnv("SMS_BASE_URL", "http://sms.etc.uz:8084"),
		SMSLogin:    getEnv("SMS_LOGIN", ""),
		SMSPassword: getEnv("SMS_PASSWORD", ""),
		SMSSender:   getEnv("SMS_SENDER", "Amity"),

		// OTP Settings
		OTPLength:      getEnvInt("OTP_LENGTH", 6),
		OTPExpiresMins: getEnvInt("OTP_EXPIRES_MINS", 5),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if c.SMSLogin == "" {
		return fmt.Errorf("SMS_LOGIN is required")
	}
	if c.SMSPassword == "" {
		return fmt.Errorf("SMS_PASSWORD is required")
	}
	return nil
}

func (c *Config) GetPostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresDB,
		c.PostgresSSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
