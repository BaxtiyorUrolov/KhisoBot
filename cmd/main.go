// cmd/bot/main.go
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"khisobot/config"
	"khisobot/internal/container"
	"khisobot/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.NewSlogLogger()

	log.Info("🚀 Starting KhisoBot...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("❌ Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("📋 Configuration loaded",
		slog.String("environment", cfg.Environment))

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize container with all dependencies
	appContainer, err := container.NewContainer(ctx, cfg, log)
	if err != nil {
		log.Error("❌ Failed to initialize container", slog.Any("error", err))
		os.Exit(1)
	}

	defer func() {
		log.Info("🔴 Closing application...")
		if err := appContainer.Close(); err != nil {
			log.Error("❌ Failed to close container", slog.Any("error", err))
		}
	}()

	// Get bot and handler
	bot := appContainer.GetBot()
	handler := appContainer.GetBotHandler()

	// Configure updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info("✅ Bot is running and listening for updates...")

	// Main loop
	for {
		select {
		case update := <-updates:
			go func(upd tgbotapi.Update) {
				handler.HandleUpdate(ctx, upd)
			}(update)

		case sig := <-sigChan:
			log.Info("📴 Received shutdown signal",
				slog.String("signal", sig.String()))

			// Stop receiving updates
			bot.StopReceivingUpdates()

			log.Info("✅ Graceful shutdown completed")
			return

		case <-ctx.Done():
			log.Info("📴 Context cancelled, shutting down...")
			return
		}
	}
}
