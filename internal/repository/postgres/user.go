// internal/repository/postgres/user.go
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

type UserRepository struct {
	db *storage.Storage
}

func NewUserRepository(db *storage.Storage) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (telegram_id, username, language_code, state, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id`

	now := time.Now()
	err := r.db.Pool.QueryRow(ctx, query,
		user.TelegramID,
		user.Username,
		user.LanguageCode,
		domain.StateWaitFullName,
		now,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	query := `
		SELECT id, telegram_id, username, language_code, first_name, last_name,
		       region, district, school, grade, phone, is_verified, state, created_at, updated_at
		FROM users
		WHERE telegram_id = $1`

	var user domain.User
	var firstName, lastName, region, district, school, phone sql.NullString
	var grade sql.NullInt32

	err := r.db.Pool.QueryRow(ctx, query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.LanguageCode,
		&firstName,
		&lastName,
		&region,
		&district,
		&school,
		&grade,
		&phone,
		&user.IsVerified,
		&user.State,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by telegram id: %w", err)
	}

	user.FirstName = firstName.String
	user.LastName = lastName.String
	user.Region = region.String
	user.District = district.String
	user.School = school.String
	user.Phone = phone.String
	user.Grade = int(grade.Int32)

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = $2, language_code = $3, first_name = $4, last_name = $5,
		    region = $6, district = $7, school = $8, grade = $9, phone = $10,
		    is_verified = $11, state = $12, updated_at = $13
		WHERE telegram_id = $1`

	_, err := r.db.Pool.Exec(ctx, query,
		user.TelegramID,
		user.Username,
		user.LanguageCode,
		user.FirstName,
		user.LastName,
		user.Region,
		user.District,
		user.School,
		user.Grade,
		user.Phone,
		user.IsVerified,
		user.State,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateState(ctx context.Context, telegramID int64, state string) error {
	query := `UPDATE users SET state = $2, updated_at = $3 WHERE telegram_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, telegramID, state, time.Now())
	if err != nil {
		return fmt.Errorf("update user state: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateFullName(ctx context.Context, telegramID int64, firstName, lastName string) error {
	query := `UPDATE users SET first_name = $2, last_name = $3, updated_at = $4 WHERE telegram_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, telegramID, firstName, lastName, time.Now())
	if err != nil {
		return fmt.Errorf("update full name: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateLocation(ctx context.Context, telegramID int64, region, district, school string) error {
	query := `UPDATE users SET region = $2, district = $3, school = $4, updated_at = $5 WHERE telegram_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, telegramID, region, district, school, time.Now())
	if err != nil {
		return fmt.Errorf("update location: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateGrade(ctx context.Context, telegramID int64, grade int) error {
	query := `UPDATE users SET grade = $2, updated_at = $3 WHERE telegram_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, telegramID, grade, time.Now())
	if err != nil {
		return fmt.Errorf("update grade: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdatePhone(ctx context.Context, telegramID int64, phone string) error {
	query := `UPDATE users SET phone = $2, updated_at = $3 WHERE telegram_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, telegramID, phone, time.Now())
	if err != nil {
		return fmt.Errorf("update phone: %w", err)
	}
	return nil
}

func (r *UserRepository) GetAllVerified(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, telegram_id, username, language_code, first_name, last_name,
		       region, district, school, grade, phone, is_verified, state, created_at, updated_at
		FROM users
		WHERE is_verified = TRUE
		ORDER BY created_at DESC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all verified users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		var firstName, lastName, region, district, school, phone sql.NullString
		var grade sql.NullInt32

		if err := rows.Scan(
			&user.ID, &user.TelegramID, &user.Username, &user.LanguageCode,
			&firstName, &lastName, &region, &district, &school, &grade, &phone,
			&user.IsVerified, &user.State, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}

		user.FirstName = firstName.String
		user.LastName = lastName.String
		user.Region = region.String
		user.District = district.String
		user.School = school.String
		user.Phone = phone.String
		user.Grade = int(grade.Int32)

		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) GetStats(ctx context.Context) (*domain.Stats, error) {
	var stats domain.Stats

	// Total users
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("get total users: %w", err)
	}

	// Verified users
	err = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE is_verified = TRUE`).Scan(&stats.VerifiedUsers)
	if err != nil {
		return nil, fmt.Errorf("get verified users: %w", err)
	}

	// Today users
	err = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE`).Scan(&stats.TodayUsers)
	if err != nil {
		return nil, fmt.Errorf("get today users: %w", err)
	}

	// Total channels
	err = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM channels WHERE is_active = TRUE`).Scan(&stats.TotalChannels)
	if err != nil {
		return nil, fmt.Errorf("get total channels: %w", err)
	}

	return &stats, nil
}
