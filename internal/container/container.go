// internal/container/container.go
package container

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"khisobot/config"
	"khisobot/internal/bot"
	"khisobot/internal/domain"
	"khisobot/internal/repository/postgres"
	"khisobot/internal/service"
	"khisobot/pkg/storage"
)

type Container struct {
	// Infra
	storage *storage.Storage
	logger  *slog.Logger
	config  *config.Config
	bot     *tgbotapi.BotAPI

	// Repos
	userRepo    domain.UserRepository
	otpRepo     domain.OTPRepository
	adminRepo   domain.AdminRepository
	channelRepo domain.ChannelRepository

	// Services
	userService *service.UserService
	otpService  *service.OTPService
	smsService  *service.SMSService

	// Bot Handler
	botHandler *bot.Handler
}

func NewContainer(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Container, error) {
	c := &Container{
		config: cfg,
		logger: logger,
	}

	if err := c.initStorage(ctx); err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}

	if err := c.initBot(); err != nil {
		return nil, fmt.Errorf("init bot: %w", err)
	}

	c.initRepositories()
	c.initServices()
	c.initBotHandler()

	logger.Info("✅ Container initialized successfully")
	return c, nil
}

func (c *Container) initStorage(ctx context.Context) error {
	st, err := storage.NewPostgresStorage(ctx, c.config)
	if err != nil {
		return fmt.Errorf("postgres storage: %w", err)
	}
	c.storage = st
	c.logger.Info("✅ Database connected")
	return nil
}

func (c *Container) initBot() error {
	botAPI, err := tgbotapi.NewBotAPI(c.config.TelegramBotToken)
	if err != nil {
		return fmt.Errorf("telegram bot: %w", err)
	}

	botAPI.Debug = c.config.Environment != "production"
	c.bot = botAPI
	c.logger.Info("✅ Telegram bot initialized",
		slog.String("username", botAPI.Self.UserName))
	return nil
}

func (c *Container) initRepositories() {
	c.userRepo = postgres.NewUserRepository(c.storage)
	c.otpRepo = postgres.NewOTPRepository(c.storage)
	c.adminRepo = postgres.NewAdminRepository(c.storage)
	c.channelRepo = postgres.NewChannelRepository(c.storage)
	c.logger.Info("✅ Repositories initialized")
}

func (c *Container) initServices() {
	c.smsService = service.NewSMSService(c.config, c.logger)
	c.userService = service.NewUserService(c.userRepo, c.logger)
	c.otpService = service.NewOTPService(c.otpRepo, c.smsService, c.config, c.logger)
	c.logger.Info("✅ Services initialized")
}

func (c *Container) initBotHandler() {
	c.botHandler = bot.NewHandler(
		c.bot,
		c.userService,
		c.otpService,
		c.adminRepo,
		c.channelRepo,
		c.logger,
	)
	c.logger.Info("✅ Bot handler initialized")
}

func (c *Container) GetBot() *tgbotapi.BotAPI {
	return c.bot
}

func (c *Container) GetBotHandler() *bot.Handler {
	return c.botHandler
}

func (c *Container) Close() error {
	c.logger.Info("🔴 Closing container resources...")

	if c.storage != nil {
		if err := c.storage.Close(); err != nil {
			c.logger.Error("❌ Failed to close storage", slog.Any("error", err))
			return err
		}
	}

	c.logger.Info("✅ Container closed successfully")
	return nil
}
