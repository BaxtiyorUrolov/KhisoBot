// internal/domain/user.go
package domain

import (
	"context"
	"time"
)

// User states
const (
	StateStart        = "start"
	StateWaitFullName = "wait_full_name"
	StateWaitLocation = "wait_location"
	StateWaitGrade    = "wait_grade"
	StateWaitPhone    = "wait_phone"
	StateWaitOTP      = "wait_otp"
	StateRegistered   = "registered"
)

// Admin states
const (
	AdminStateNone          = ""
	AdminStateWaitChannel   = "wait_channel"
	AdminStateWaitBroadcast = "wait_broadcast"
)

type User struct {
	ID           int64     `db:"id"`
	TelegramID   int64     `db:"telegram_id"`
	Username     string    `db:"username"`
	LanguageCode string    `db:"language_code"`
	FirstName    string    `db:"first_name"`
	LastName     string    `db:"last_name"`
	Region       string    `db:"region"`
	District     string    `db:"district"`
	School       string    `db:"school"`
	Grade        int       `db:"grade"`
	Phone        string    `db:"phone"`
	IsVerified   bool      `db:"is_verified"`
	State        string    `db:"state"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type OTPCode struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Phone     string    `db:"phone"`
	Code      string    `db:"code"`
	MessageID string    `db:"message_id"`
	IsUsed    bool      `db:"is_used"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type Admin struct {
	ID         int64     `db:"id"`
	TelegramID int64     `db:"telegram_id"`
	Username   string    `db:"username"`
	State      string    `db:"-"` // Not in DB, used in memory
	CreatedAt  time.Time `db:"created_at"`
}

type Channel struct {
	ID              int64     `db:"id"`
	ChannelID       int64     `db:"channel_id"`
	ChannelUsername string    `db:"channel_username"`
	Title           string    `db:"title"`
	IsActive        bool      `db:"is_active"`
	CreatedAt       time.Time `db:"created_at"`
}

type Stats struct {
	TotalUsers    int64
	VerifiedUsers int64
	TodayUsers    int64
	TotalChannels int64
}

// UserRepository interface
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	Update(ctx context.Context, user *User) error
	UpdateState(ctx context.Context, telegramID int64, state string) error
	UpdateFullName(ctx context.Context, telegramID int64, firstName, lastName string) error
	UpdateLocation(ctx context.Context, telegramID int64, region, district, school string) error
	UpdateGrade(ctx context.Context, telegramID int64, grade int) error
	UpdatePhone(ctx context.Context, telegramID int64, phone string) error
	GetAllVerified(ctx context.Context) ([]User, error)
	GetStats(ctx context.Context) (*Stats, error)
}

// OTPRepository interface
type OTPRepository interface {
	Create(ctx context.Context, otp *OTPCode) error
	GetLatestByPhone(ctx context.Context, phone string) (*OTPCode, error)
	MarkAsUsed(ctx context.Context, id int64) error
	GetByPhoneAndCode(ctx context.Context, phone, code string) (*OTPCode, error)
}

// AdminRepository interface
type AdminRepository interface {
	IsAdmin(ctx context.Context, telegramID int64) (bool, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*Admin, error)
}

// ChannelRepository interface
type ChannelRepository interface {
	Create(ctx context.Context, channel *Channel) error
	GetAll(ctx context.Context) ([]Channel, error)
	GetActive(ctx context.Context) ([]Channel, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Channel, error)
}

// UserService interface
type UserService interface {
	GetOrCreateUser(ctx context.Context, telegramID int64, username, langCode string) (*User, error)
	UpdateUserState(ctx context.Context, telegramID int64, state string) error
	UpdateFullName(ctx context.Context, telegramID int64, firstName, lastName string) error
	UpdateLocation(ctx context.Context, telegramID int64, region, district, school string) error
	UpdateGrade(ctx context.Context, telegramID int64, grade int) error
	UpdatePhone(ctx context.Context, telegramID int64, phone string) error
	GetUser(ctx context.Context, telegramID int64) (*User, error)
	VerifyUser(ctx context.Context, telegramID int64) error
	GetAllVerified(ctx context.Context) ([]User, error)
	GetStats(ctx context.Context) (*Stats, error)
}

// OTPService interface
type OTPService interface {
	GenerateAndSendOTP(ctx context.Context, userID int64, phone string) error
	VerifyOTP(ctx context.Context, phone, code string) (bool, error)
}
