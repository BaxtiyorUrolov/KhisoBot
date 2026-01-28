// internal/repository/postgres/otp.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"khisobot/internal/domain"
	"khisobot/pkg/storage"
)

type OTPRepository struct {
	db *storage.Storage
}

func NewOTPRepository(db *storage.Storage) *OTPRepository {
	return &OTPRepository{db: db}
}

func (r *OTPRepository) Create(ctx context.Context, otp *domain.OTPCode) error {
	query := `
		INSERT INTO otp_codes (user_id, phone, code, message_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	now := time.Now()
	err := r.db.Pool.QueryRow(ctx, query,
		otp.UserID,
		otp.Phone,
		otp.Code,
		otp.MessageID,
		otp.ExpiresAt,
		now,
	).Scan(&otp.ID)

	if err != nil {
		return fmt.Errorf("create otp: %w", err)
	}

	otp.CreatedAt = now
	return nil
}

func (r *OTPRepository) GetLatestByPhone(ctx context.Context, phone string) (*domain.OTPCode, error) {
	query := `
		SELECT id, user_id, phone, code, message_id, is_used, expires_at, created_at
		FROM otp_codes
		WHERE phone = $1 AND is_used = FALSE AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1`

	var otp domain.OTPCode
	var msgID sql.NullString

	err := r.db.Pool.QueryRow(ctx, query, phone).Scan(
		&otp.ID,
		&otp.UserID,
		&otp.Phone,
		&otp.Code,
		&msgID,
		&otp.IsUsed,
		&otp.ExpiresAt,
		&otp.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest otp: %w", err)
	}

	otp.MessageID = msgID.String
	return &otp, nil
}

func (r *OTPRepository) GetByPhoneAndCode(ctx context.Context, phone, code string) (*domain.OTPCode, error) {
	query := `
		SELECT id, user_id, phone, code, message_id, is_used, expires_at, created_at
		FROM otp_codes
		WHERE phone = $1 AND code = $2 AND is_used = FALSE AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1`

	var otp domain.OTPCode
	var msgID sql.NullString

	err := r.db.Pool.QueryRow(ctx, query, phone, code).Scan(
		&otp.ID,
		&otp.UserID,
		&otp.Phone,
		&otp.Code,
		&msgID,
		&otp.IsUsed,
		&otp.ExpiresAt,
		&otp.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get otp by phone and code: %w", err)
	}

	otp.MessageID = msgID.String
	return &otp, nil
}

func (r *OTPRepository) MarkAsUsed(ctx context.Context, id int64) error {
	query := `UPDATE otp_codes SET is_used = TRUE WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark otp as used: %w", err)
	}
	return nil
}
