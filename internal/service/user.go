// internal/service/user.go
package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"khisobot/config"
	"khisobot/internal/domain"
)

type UserService struct {
	userRepo domain.UserRepository
	logger   *slog.Logger
}

func NewUserService(userRepo domain.UserRepository, logger *slog.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *UserService) GetOrCreateUser(ctx context.Context, telegramID int64, username, langCode string) (*domain.User, error) {
	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if user != nil {
		s.logger.Info("👤 Existing user found",
			slog.Int64("telegram_id", telegramID),
			slog.String("state", user.State))
		return user, nil
	}

	if langCode == "" || (langCode != "uz" && langCode != "ru" && langCode != "en") {
		langCode = "uz"
	}

	newUser := &domain.User{
		TelegramID:   telegramID,
		Username:     username,
		LanguageCode: langCode,
		State:        domain.StateWaitFullName,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.logger.Info("✅ New user created",
		slog.Int64("telegram_id", telegramID),
		slog.String("language", langCode))

	return newUser, nil
}

func (s *UserService) GetUser(ctx context.Context, telegramID int64) (*domain.User, error) {
	return s.userRepo.GetByTelegramID(ctx, telegramID)
}

func (s *UserService) UpdateUserState(ctx context.Context, telegramID int64, state string) error {
	return s.userRepo.UpdateState(ctx, telegramID, state)
}

func (s *UserService) UpdateFullName(ctx context.Context, telegramID int64, firstName, lastName string) error {
	return s.userRepo.UpdateFullName(ctx, telegramID, firstName, lastName)
}

func (s *UserService) UpdateLocation(ctx context.Context, telegramID int64, region, district, school string) error {
	return s.userRepo.UpdateLocation(ctx, telegramID, region, district, school)
}

func (s *UserService) UpdateGrade(ctx context.Context, telegramID int64, grade int) error {
	return s.userRepo.UpdateGrade(ctx, telegramID, grade)
}

func (s *UserService) UpdatePhone(ctx context.Context, telegramID int64, phone string) error {
	return s.userRepo.UpdatePhone(ctx, telegramID, phone)
}

func (s *UserService) VerifyUser(ctx context.Context, telegramID int64) error {
	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	user.IsVerified = true
	user.State = domain.StateRegistered

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) GetAllVerified(ctx context.Context) ([]domain.User, error) {
	return s.userRepo.GetAllVerified(ctx)
}

func (s *UserService) GetStats(ctx context.Context) (*domain.Stats, error) {
	return s.userRepo.GetStats(ctx)
}

// OTPService implementation
type OTPService struct {
	otpRepo    domain.OTPRepository
	smsService *SMSService
	cfg        *config.Config
	logger     *slog.Logger
}

func NewOTPService(otpRepo domain.OTPRepository, smsService *SMSService, cfg *config.Config, logger *slog.Logger) *OTPService {
	return &OTPService{
		otpRepo:    otpRepo,
		smsService: smsService,
		cfg:        cfg,
		logger:     logger,
	}
}

func (s *OTPService) GenerateAndSendOTP(ctx context.Context, userID int64, phone string) error {
	code, err := generateOTPCode(s.cfg.OTPLength)
	if err != nil {
		return fmt.Errorf("generate otp code: %w", err)
	}

	message := fmt.Sprintf("Sizning tasdiqlash kodingiz: %s\nKod %d daqiqa ichida amal qiladi.",
		code, s.cfg.OTPExpiresMins)

	messageID, err := s.smsService.SendSMS(ctx, phone, message)
	if err != nil {
		return fmt.Errorf("send sms: %w", err)
	}

	otp := &domain.OTPCode{
		UserID:    userID,
		Phone:     phone,
		Code:      code,
		MessageID: messageID,
		ExpiresAt: time.Now().Add(time.Duration(s.cfg.OTPExpiresMins) * time.Minute),
	}

	if err := s.otpRepo.Create(ctx, otp); err != nil {
		return fmt.Errorf("save otp: %w", err)
	}

	s.logger.Info("📱 OTP sent",
		slog.String("phone", phone),
		slog.String("message_id", messageID))

	return nil
}

func (s *OTPService) VerifyOTP(ctx context.Context, phone, code string) (bool, error) {
	otp, err := s.otpRepo.GetByPhoneAndCode(ctx, phone, code)
	if err != nil {
		return false, fmt.Errorf("get otp: %w", err)
	}

	if otp == nil {
		s.logger.Warn("❌ Invalid OTP code", slog.String("phone", phone))
		return false, nil
	}

	if err := s.otpRepo.MarkAsUsed(ctx, otp.ID); err != nil {
		return false, fmt.Errorf("mark otp as used: %w", err)
	}

	s.logger.Info("✅ OTP verified", slog.String("phone", phone))
	return true, nil
}

func generateOTPCode(length int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, length)

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[n.Int64()]
	}

	return string(code), nil
}
