// internal/repository/postgres/admin.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"khisobot/internal/domain"
	"khisobot/pkg/storage"
)

type AdminRepository struct {
	db *storage.Storage
}

func NewAdminRepository(db *storage.Storage) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) IsAdmin(ctx context.Context, telegramID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM admins WHERE telegram_id = $1)`
	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, telegramID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check admin: %w", err)
	}
	return exists, nil
}

func (r *AdminRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Admin, error) {
	query := `SELECT id, telegram_id, username, created_at FROM admins WHERE telegram_id = $1`

	var admin domain.Admin
	var username sql.NullString

	err := r.db.Pool.QueryRow(ctx, query, telegramID).Scan(
		&admin.ID,
		&admin.TelegramID,
		&username,
		&admin.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get admin: %w", err)
	}

	admin.Username = username.String
	return &admin, nil
}

// ChannelRepository
type ChannelRepository struct {
	db *storage.Storage
}

func NewChannelRepository(db *storage.Storage) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) Create(ctx context.Context, channel *domain.Channel) error {
	query := `
		INSERT INTO channels (channel_id, channel_username, title, is_active)
		VALUES ($1, $2, $3, TRUE)
		RETURNING id, created_at`

	err := r.db.Pool.QueryRow(ctx, query,
		channel.ChannelID,
		channel.ChannelUsername,
		channel.Title,
	).Scan(&channel.ID, &channel.CreatedAt)

	if err != nil {
		return fmt.Errorf("create channel: %w", err)
	}
	return nil
}

func (r *ChannelRepository) GetAll(ctx context.Context) ([]domain.Channel, error) {
	query := `SELECT id, channel_id, channel_username, title, is_active, created_at FROM channels ORDER BY created_at DESC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all channels: %w", err)
	}
	defer rows.Close()

	var channels []domain.Channel
	for rows.Next() {
		var ch domain.Channel
		var title sql.NullString
		var channelID sql.NullInt64

		if err := rows.Scan(&ch.ID, &channelID, &ch.ChannelUsername, &title, &ch.IsActive, &ch.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		ch.ChannelID = channelID.Int64
		ch.Title = title.String
		channels = append(channels, ch)
	}

	return channels, nil
}

func (r *ChannelRepository) GetActive(ctx context.Context) ([]domain.Channel, error) {
	query := `SELECT id, channel_id, channel_username, title, is_active, created_at FROM channels WHERE is_active = TRUE`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get active channels: %w", err)
	}
	defer rows.Close()

	var channels []domain.Channel
	for rows.Next() {
		var ch domain.Channel
		var title sql.NullString
		var channelID sql.NullInt64

		if err := rows.Scan(&ch.ID, &channelID, &ch.ChannelUsername, &title, &ch.IsActive, &ch.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		ch.ChannelID = channelID.Int64
		ch.Title = title.String
		channels = append(channels, ch)
	}

	return channels, nil
}

func (r *ChannelRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM channels WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}
	return nil
}

func (r *ChannelRepository) GetByID(ctx context.Context, id int64) (*domain.Channel, error) {
	query := `SELECT id, channel_id, channel_username, title, is_active, created_at FROM channels WHERE id = $1`

	var ch domain.Channel
	var title sql.NullString
	var channelID sql.NullInt64

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&ch.ID, &channelID, &ch.ChannelUsername, &title, &ch.IsActive, &ch.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get channel by id: %w", err)
	}

	ch.ChannelID = channelID.Int64
	ch.Title = title.String
	return &ch, nil
}
