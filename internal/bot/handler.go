// internal/bot/handler.go
package bot

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"khisobot/internal/domain"
	"khisobot/internal/service"
	"khisobot/pkg/i18n"
)

var phoneRegex = regexp.MustCompile(`^998[0-9]{9}$`)

const (
	CallbackLogin       = "login"
	CallbackRegister    = "register"
	CallbackCheckSub    = "check_sub"
	CallbackResendOTP   = "resend_otp"
	CallbackAdminStats  = "admin_stats"
	CallbackAdminAdd    = "admin_add_channel"
	CallbackAdminRemove = "admin_remove_channel"
	CallbackAdminExport = "admin_export"
	CallbackAdminBack   = "admin_back"
	CallbackDelChannel  = "del_ch_"
)

type Handler struct {
	bot         *tgbotapi.BotAPI
	userService *service.UserService
	otpService  *service.OTPService
	adminRepo   domain.AdminRepository
	channelRepo domain.ChannelRepository
	logger      *slog.Logger

	// Admin states (in memory)
	adminStates map[int64]string
	mu          sync.RWMutex
}

func NewHandler(
	bot *tgbotapi.BotAPI,
	userService *service.UserService,
	otpService *service.OTPService,
	adminRepo domain.AdminRepository,
	channelRepo domain.ChannelRepository,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		bot:         bot,
		userService: userService,
		otpService:  otpService,
		adminRepo:   adminRepo,
		channelRepo: channelRepo,
		logger:      logger,
		adminStates: make(map[int64]string),
	}
}

func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message != nil {
		if update.Message.IsCommand() {
			h.handleCommand(ctx, update.Message)
			return
		}
		h.handleMessage(ctx, update.Message)
		return
	}

	if update.CallbackQuery != nil {
		h.handleCallback(ctx, update.CallbackQuery)
	}
}

func (h *Handler) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		h.handleStart(ctx, msg)
	case "resend":
		h.handleResendOTP(ctx, msg)
	case "profile":
		h.handleProfile(ctx, msg)
	case "admin":
		h.handleAdmin(ctx, msg)
	}
}

func (h *Handler) handleStart(ctx context.Context, msg *tgbotapi.Message) {
	langCode := msg.From.LanguageCode
	if langCode == "" {
		langCode = "uz"
	}

	// Check subscription first
	if !h.checkSubscription(ctx, msg.From.ID, msg.Chat.ID, langCode) {
		return
	}

	user, err := h.userService.GetOrCreateUser(ctx, msg.From.ID, msg.From.UserName, langCode)
	if err != nil {
		h.logger.Error("❌ Failed to get/create user", slog.Any("error", err))
		h.sendMessage(msg.Chat.ID, i18n.Get(langCode).Error)
		return
	}

	if user.State == domain.StateRegistered && user.IsVerified {
		h.sendMainMenu(msg.Chat.ID, user.LanguageCode)
		return
	}

	msgs := i18n.Get(user.LanguageCode)
	h.sendMessageHTML(msg.Chat.ID, msgs.Welcome+"\n\n"+msgs.AskFullName)
}

func (h *Handler) checkSubscription(ctx context.Context, userID int64, chatID int64, lang string) bool {
	channels, err := h.channelRepo.GetActive(ctx)
	if err != nil || len(channels) == 0 {
		return true
	}

	var notSubscribed []domain.Channel
	for _, ch := range channels {
		member, err := h.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID:             0,
				SuperGroupUsername: "@" + ch.ChannelUsername,
				UserID:             userID,
			},
		})
		if err != nil || member.Status == "left" || member.Status == "kicked" {
			notSubscribed = append(notSubscribed, ch)
		}
	}

	if len(notSubscribed) == 0 {
		return true
	}

	// Build keyboard with channel buttons
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, ch := range notSubscribed {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📢 "+ch.ChannelUsername, "https://t.me/"+ch.ChannelUsername),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(i18n.Get(lang).BtnCheckSub, CallbackCheckSub),
	))

	msg := tgbotapi.NewMessage(chatID, i18n.Get(lang).MustSubscribe)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.bot.Send(msg)

	return false
}

func (h *Handler) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	// Check admin state first
	h.mu.RLock()
	adminState := h.adminStates[msg.From.ID]
	h.mu.RUnlock()

	if adminState == domain.AdminStateWaitChannel {
		h.handleAddChannel(ctx, msg)
		return
	}

	user, err := h.userService.GetUser(ctx, msg.From.ID)
	if err != nil || user == nil {
		h.handleStart(ctx, msg)
		return
	}

	// Contact yuborilgan bo'lsa (telefon raqam ulashish)
	if msg.Contact != nil && user.State == domain.StateWaitPhone {
		h.handleContact(ctx, msg, user)
		return
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}

	switch user.State {
	case domain.StateWaitFullName:
		h.handleFullName(ctx, msg, user, text)
	case domain.StateWaitLocation:
		h.handleLocation(ctx, msg, user, text)
	case domain.StateWaitGrade:
		h.handleGrade(ctx, msg, user, text)
	case domain.StateWaitPhone:
		h.handlePhone(ctx, msg, user, text)
	case domain.StateWaitOTP:
		h.handleOTPInput(ctx, msg, user, text)
	case domain.StateRegistered:
		h.sendMainMenu(msg.Chat.ID, user.LanguageCode)
	}
}

func (h *Handler) handleContact(ctx context.Context, msg *tgbotapi.Message, user *domain.User) {
	// Kontaktdan telefon raqamni olish
	phone := msg.Contact.PhoneNumber
	phone = strings.NewReplacer(" ", "", "+", "", "-", "", "(", "", ")", "").Replace(phone)

	// Agar 998 bilan boshlanmasa, qo'shish
	if !strings.HasPrefix(phone, "998") && len(phone) == 9 {
		phone = "998" + phone
	}

	if !phoneRegex.MatchString(phone) {
		h.sendMessageHTML(msg.Chat.ID, i18n.Get(user.LanguageCode).InvalidPhone)
		return
	}

	if err := h.userService.UpdatePhone(ctx, user.TelegramID, phone); err != nil {
		h.sendMessage(msg.Chat.ID, i18n.Get(user.LanguageCode).Error)
		return
	}

	updatedUser, _ := h.userService.GetUser(ctx, user.TelegramID)

	if err := h.otpService.GenerateAndSendOTP(ctx, updatedUser.ID, phone); err != nil {
		h.logger.Error("❌ Failed to send OTP", slog.Any("error", err))
		h.sendMessage(msg.Chat.ID, i18n.Get(user.LanguageCode).Error)
		return
	}

	h.userService.UpdateUserState(ctx, user.TelegramID, domain.StateWaitOTP)

	// Keyboard'ni olib tashlash
	removeKeyboard := tgbotapi.NewRemoveKeyboard(true)
	msgRemove := tgbotapi.NewMessage(msg.Chat.ID, "✅")
	msgRemove.ReplyMarkup = removeKeyboard
	h.bot.Send(msgRemove)

	maskedPhone := phone[:6] + "****" + phone[len(phone)-2:]
	h.sendOTPMessage(msg.Chat.ID, user.LanguageCode, maskedPhone)
}

func (h *Handler) handleFullName(ctx context.Context, msg *tgbotapi.Message, user *domain.User, text string) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		h.sendMessageHTML(msg.Chat.ID, i18n.Get(user.LanguageCode).InvalidFullName)
		return
	}

	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")

	if err := h.userService.UpdateFullName(ctx, user.TelegramID, firstName, lastName); err != nil {
		h.logger.Error("❌ Failed to update full name", slog.Any("error", err))
		h.sendMessage(msg.Chat.ID, i18n.Get(user.LanguageCode).Error)
		return
	}

	if err := h.userService.UpdateUserState(ctx, user.TelegramID, domain.StateWaitLocation); err != nil {
		return
	}

	h.sendMessageHTML(msg.Chat.ID, i18n.Get(user.LanguageCode).AskLocation)
}

func (h *Handler) handleLocation(ctx context.Context, msg *tgbotapi.Message, user *domain.User, text string) {
	parts := strings.Split(text, ",")
	if len(parts) < 3 {
		h.sendMessageHTML(msg.Chat.ID, i18n.Get(user.LanguageCode).InvalidLocation)
		return
	}

	region := strings.TrimSpace(parts[0])
	district := strings.TrimSpace(parts[1])
	school := strings.TrimSpace(strings.Join(parts[2:], ","))

	if region == "" || district == "" || school == "" {
		h.sendMessageHTML(msg.Chat.ID, i18n.Get(user.LanguageCode).InvalidLocation)
		return
	}

	if err := h.userService.UpdateLocation(ctx, user.TelegramID, region, district, school); err != nil {
		h.sendMessage(msg.Chat.ID, i18n.Get(user.LanguageCode).Error)
		return
	}

	if err := h.userService.UpdateUserState(ctx, user.TelegramID, domain.StateWaitGrade); err != nil {
		return
	}

	h.sendMessageHTML(msg.Chat.ID, i18n.Get(user.LanguageCode).AskGrade)
}

func (h *Handler) handleGrade(ctx context.Context, msg *tgbotapi.Message, user *domain.User, text string) {
	grade, err := strconv.Atoi(text)
	if err != nil || grade < 1 || grade > 11 {
		h.sendMessageHTML(msg.Chat.ID, i18n.Get(user.LanguageCode).InvalidGrade)
		return
	}

	if err := h.userService.UpdateGrade(ctx, user.TelegramID, grade); err != nil {
		h.sendMessage(msg.Chat.ID, i18n.Get(user.LanguageCode).Error)
		return
	}

	if err := h.userService.UpdateUserState(ctx, user.TelegramID, domain.StateWaitPhone); err != nil {
		return
	}

	// Telefon so'rash - contact button bilan
	h.sendPhoneRequest(msg.Chat.ID, user.LanguageCode)
}

func (h *Handler) sendPhoneRequest(chatID int64, langCode string) {
	msgs := i18n.Get(langCode)

	// ReplyKeyboard bilan "Share Contact" tugmasi
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Telefon raqamni ulashish"),
		),
	)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true

	msg := tgbotapi.NewMessage(chatID, msgs.AskPhone)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *Handler) handlePhone(ctx context.Context, msg *tgbotapi.Message, user *domain.User, text string) {
	phone := strings.NewReplacer(" ", "", "+", "", "-", "", "(", "", ")", "").Replace(text)

	if !phoneRegex.MatchString(phone) {
		h.sendMessageHTML(msg.Chat.ID, i18n.Get(user.LanguageCode).InvalidPhone)
		return
	}

	if err := h.userService.UpdatePhone(ctx, user.TelegramID, phone); err != nil {
		h.sendMessage(msg.Chat.ID, i18n.Get(user.LanguageCode).Error)
		return
	}

	updatedUser, _ := h.userService.GetUser(ctx, user.TelegramID)

	if err := h.otpService.GenerateAndSendOTP(ctx, updatedUser.ID, phone); err != nil {
		h.logger.Error("❌ Failed to send OTP", slog.Any("error", err))
		h.sendMessage(msg.Chat.ID, i18n.Get(user.LanguageCode).Error)
		return
	}

	h.userService.UpdateUserState(ctx, user.TelegramID, domain.StateWaitOTP)

	// Keyboard'ni olib tashlash
	removeKeyboard := tgbotapi.NewRemoveKeyboard(true)
	msgRemove := tgbotapi.NewMessage(msg.Chat.ID, "✅")
	msgRemove.ReplyMarkup = removeKeyboard
	h.bot.Send(msgRemove)

	maskedPhone := phone[:6] + "****" + phone[len(phone)-2:]
	h.sendOTPMessage(msg.Chat.ID, user.LanguageCode, maskedPhone)
}

func (h *Handler) handleOTPInput(ctx context.Context, message *tgbotapi.Message, user *domain.User, text string) {
	currentUser, _ := h.userService.GetUser(ctx, user.TelegramID)
	if currentUser == nil || currentUser.Phone == "" {
		h.sendMessage(message.Chat.ID, i18n.Get(user.LanguageCode).Error)
		return
	}

	valid, _ := h.otpService.VerifyOTP(ctx, currentUser.Phone, strings.TrimSpace(text))
	if !valid {
		// Noto'g'ri kod - qayta yuborish tugmasi bilan
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(i18n.Get(user.LanguageCode).ResendOTP, CallbackResendOTP),
			),
		)
		msg := tgbotapi.NewMessage(message.Chat.ID, i18n.Get(user.LanguageCode).InvalidOTP)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)
		return
	}

	h.userService.VerifyUser(ctx, user.TelegramID)

	finalUser, _ := h.userService.GetUser(ctx, user.TelegramID)
	successMsg := fmt.Sprintf(i18n.Get(user.LanguageCode).RegistrationDone,
		finalUser.FirstName, finalUser.LastName, finalUser.Region,
		finalUser.District, finalUser.School, finalUser.Grade, finalUser.Phone)
	h.sendMessageHTML(message.Chat.ID, successMsg)

	h.sendMainMenu(message.Chat.ID, user.LanguageCode)
}

func (h *Handler) handleResendOTP(ctx context.Context, msg *tgbotapi.Message) {
	user, _ := h.userService.GetUser(ctx, msg.From.ID)
	if user == nil || user.Phone == "" || user.State != domain.StateWaitOTP {
		return
	}

	if err := h.otpService.GenerateAndSendOTP(ctx, user.ID, user.Phone); err != nil {
		h.logger.Error("❌ Failed to resend OTP", slog.Any("error", err))
		h.sendMessage(msg.Chat.ID, i18n.Get(user.LanguageCode).Error)
		return
	}

	maskedPhone := user.Phone[:6] + "****" + user.Phone[len(user.Phone)-2:]
	h.sendOTPMessage(msg.Chat.ID, user.LanguageCode, maskedPhone)
}

func (h *Handler) handleResendOTPCallback(ctx context.Context, chatID int64, userID int64) {
	user, _ := h.userService.GetUser(ctx, userID)
	if user == nil || user.Phone == "" || user.State != domain.StateWaitOTP {
		return
	}

	if err := h.otpService.GenerateAndSendOTP(ctx, user.ID, user.Phone); err != nil {
		h.logger.Error("❌ Failed to resend OTP", slog.Any("error", err))
		h.sendMessage(chatID, i18n.Get(user.LanguageCode).Error)
		return
	}

	maskedPhone := user.Phone[:6] + "****" + user.Phone[len(user.Phone)-2:]
	h.sendOTPMessage(chatID, user.LanguageCode, maskedPhone)
}

func (h *Handler) sendOTPMessage(chatID int64, langCode string, maskedPhone string) {
	msgs := i18n.Get(langCode)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(msgs.ResendOTP, CallbackResendOTP),
		),
	)

	text := fmt.Sprintf(msgs.OTPSent, maskedPhone) + "\n\n" + msgs.AskOTP
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *Handler) handleProfile(ctx context.Context, msg *tgbotapi.Message) {
	user, _ := h.userService.GetUser(ctx, msg.From.ID)
	if user == nil || !user.IsVerified {
		h.sendMessage(msg.Chat.ID, i18n.Get("uz").NotRegistered)
		return
	}

	profileMsg := fmt.Sprintf(i18n.Get(user.LanguageCode).RegistrationDone,
		user.FirstName, user.LastName, user.Region, user.District, user.School, user.Grade, user.Phone)
	h.sendMessageHTML(msg.Chat.ID, profileMsg)
}

// ==================== ADMIN ====================

func (h *Handler) handleAdmin(ctx context.Context, msg *tgbotapi.Message) {
	isAdmin, _ := h.adminRepo.IsAdmin(ctx, msg.From.ID)
	if !isAdmin {
		return
	}

	h.sendAdminPanel(ctx, msg.Chat.ID)
}

func (h *Handler) sendAdminPanel(ctx context.Context, chatID int64) {
	stats, _ := h.userService.GetStats(ctx)

	text := fmt.Sprintf(`🔐 <b>Admin Panel</b>

📊 <b>Statistika:</b>
👥 Jami foydalanuvchilar: <b>%d</b>
✅ Tasdiqlangan: <b>%d</b>
📅 Bugun qo'shilgan: <b>%d</b>
📢 Faol kanallar: <b>%d</b>`,
		stats.TotalUsers, stats.VerifiedUsers, stats.TodayUsers, stats.TotalChannels)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Statistika", CallbackAdminStats),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Kanal qo'shish", CallbackAdminAdd),
			tgbotapi.NewInlineKeyboardButtonData("➖ Kanal o'chirish", CallbackAdminRemove),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📥 Excel yuklab olish", CallbackAdminExport),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *Handler) handleAddChannel(ctx context.Context, msg *tgbotapi.Message) {
	h.mu.Lock()
	delete(h.adminStates, msg.From.ID)
	h.mu.Unlock()

	username := strings.TrimPrefix(msg.Text, "@")
	username = strings.TrimSpace(username)

	channel := &domain.Channel{
		ChannelUsername: username,
		Title:           username,
	}

	if err := h.channelRepo.Create(ctx, channel); err != nil {
		h.sendMessage(msg.Chat.ID, "❌ Kanal qo'shishda xatolik: "+err.Error())
		return
	}

	h.sendMessage(msg.Chat.ID, fmt.Sprintf("✅ Kanal qo'shildi: @%s", username))
	h.sendAdminPanel(ctx, msg.Chat.ID)
}

func (h *Handler) handleCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	switch callback.Data {
	case CallbackCheckSub:
		lang := "uz"
		if user, _ := h.userService.GetUser(ctx, callback.From.ID); user != nil {
			lang = user.LanguageCode
		}
		if h.checkSubscription(ctx, callback.From.ID, callback.Message.Chat.ID, lang) {
			h.sendMessageHTML(callback.Message.Chat.ID, i18n.Get(lang).SubscribeSuccess)
			h.handleStartFromCallback(ctx, callback)
		}

	case CallbackResendOTP:
		h.handleResendOTPCallback(ctx, callback.Message.Chat.ID, callback.From.ID)

	case CallbackAdminStats:
		h.sendAdminPanel(ctx, callback.Message.Chat.ID)

	case CallbackAdminAdd:
		h.mu.Lock()
		h.adminStates[callback.From.ID] = domain.AdminStateWaitChannel
		h.mu.Unlock()
		h.sendMessage(callback.Message.Chat.ID, "📢 Kanal username'ini kiriting (masalan: @channel_name):")

	case CallbackAdminRemove:
		h.sendChannelList(ctx, callback.Message.Chat.ID)

	case CallbackAdminExport:
		h.exportToExcel(ctx, callback.Message.Chat.ID)

	case CallbackAdminBack:
		h.sendAdminPanel(ctx, callback.Message.Chat.ID)

	default:
		if strings.HasPrefix(callback.Data, CallbackDelChannel) {
			idStr := strings.TrimPrefix(callback.Data, CallbackDelChannel)
			id, _ := strconv.ParseInt(idStr, 10, 64)
			h.channelRepo.Delete(ctx, id)
			h.sendMessage(callback.Message.Chat.ID, "✅ Kanal o'chirildi")
			h.sendAdminPanel(ctx, callback.Message.Chat.ID)
		}
	}
}

func (h *Handler) handleStartFromCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	msg := &tgbotapi.Message{
		From: callback.From,
		Chat: callback.Message.Chat,
	}
	h.handleStart(ctx, msg)
}

func (h *Handler) sendChannelList(ctx context.Context, chatID int64) {
	channels, _ := h.channelRepo.GetAll(ctx)

	if len(channels) == 0 {
		h.sendMessage(chatID, "📢 Hozircha kanallar yo'q")
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, ch := range channels {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("🗑 @%s", ch.ChannelUsername),
				CallbackDelChannel+strconv.FormatInt(ch.ID, 10),
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Orqaga", CallbackAdminBack),
	))

	msg := tgbotapi.NewMessage(chatID, "📢 O'chirish uchun kanalni tanlang:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.bot.Send(msg)
}

func (h *Handler) exportToExcel(ctx context.Context, chatID int64) {
	users, err := h.userService.GetAllVerified(ctx)
	if err != nil {
		h.sendMessage(chatID, "❌ Xatolik: "+err.Error())
		return
	}

	f := excelize.NewFile()
	sheet := "Users"
	f.SetSheetName("Sheet1", sheet)

	// Headers
	headers := []string{"#", "Ism", "Familiya", "Viloyat", "Tuman", "Maktab", "Sinf", "Telefon", "Username", "Sana"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Data
	for i, u := range users {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), u.FirstName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), u.LastName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), u.Region)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), u.District)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), u.School)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), u.Grade)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), u.Phone)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), u.Username)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", row), u.CreatedAt.Format("02.01.2006"))
	}

	var buf bytes.Buffer
	f.Write(&buf)

	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{
		Name:  "users.xlsx",
		Bytes: buf.Bytes(),
	})
	doc.Caption = fmt.Sprintf("📊 Jami %d ta foydalanuvchi", len(users))
	h.bot.Send(doc)
}

func (h *Handler) sendMainMenu(chatID int64, langCode string) {
	msgs := i18n.Get(langCode)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(msgs.BtnLogin, "https://khiso.uz/login"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(msgs.BtnRegister, "https://khiso.uz/register"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, msgs.MainMenu)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard
	msg.DisableWebPagePreview = true
	h.bot.Send(msg)
}

func (h *Handler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	h.bot.Send(msg)
}

func (h *Handler) sendMessageHTML(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	h.bot.Send(msg)
}
