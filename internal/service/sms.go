// internal/service/sms.go
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"khisobot/config"
)

type SMSService struct {
	cfg    *config.Config
	client *http.Client
	logger *slog.Logger
}

// Request/Response structs for sms.etc.uz API
type SMSRequest struct {
	Header SMSHeader `json:"header"`
	Body   SMSBody   `json:"body"`
}

type SMSHeader struct {
	Login string `json:"login"`
	Pwd   string `json:"pwd"`
	CgPN  string `json:"CgPN"`
}

type SMSBody struct {
	MessageIDIn string `json:"message_id_in"`
	CdPN        string `json:"CdPN"`
	Text        string `json:"text"`
}

type SMSResponse struct {
	QueryCode  int    `json:"query_code"`
	QueryState string `json:"query_state"`
	MessageID  string `json:"message_id,omitempty"`
}

type SMSStatusRequest struct {
	Login       string `json:"login"`
	Pwd         string `json:"pwd"`
	MessageIDIn string `json:"message_id_in"`
}

func NewSMSService(cfg *config.Config, logger *slog.Logger) *SMSService {
	return &SMSService{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (s *SMSService) SendSMS(ctx context.Context, phone, message string) (string, error) {
	messageID := fmt.Sprintf("OTP_%d", time.Now().UnixNano())

	req := SMSRequest{
		Header: SMSHeader{
			Login: s.cfg.SMSLogin,
			Pwd:   s.cfg.SMSPassword,
			CgPN:  s.cfg.SMSSender,
		},
		Body: SMSBody{
			MessageIDIn: messageID,
			CdPN:        phone,
			Text:        message,
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal sms request: %w", err)
	}

	s.logger.Info("📤 Sending SMS",
		slog.String("phone", phone),
		slog.String("message_id", messageID))

	url := fmt.Sprintf("%s/single-sms", s.cfg.SMSBaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("send sms request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		return "", fmt.Errorf("sms service error: invalid request format")
	}

	var smsResp SMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
		return "", fmt.Errorf("decode sms response: %w", err)
	}

	switch smsResp.QueryCode {
	case 200:
		s.logger.Info("✅ SMS sent successfully",
			slog.String("phone", phone),
			slog.String("message_id", messageID))
		return messageID, nil
	case 401:
		return "", fmt.Errorf("sms auth failed: %s", smsResp.QueryState)
	case 503:
		return "", fmt.Errorf("sms service error: %s", smsResp.QueryState)
	default:
		return "", fmt.Errorf("unknown sms error: code=%d, state=%s", smsResp.QueryCode, smsResp.QueryState)
	}
}

func (s *SMSService) GetStatus(ctx context.Context, messageID string) (string, error) {
	req := SMSStatusRequest{
		Login:       s.cfg.SMSLogin,
		Pwd:         s.cfg.SMSPassword,
		MessageIDIn: messageID,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal status request: %w", err)
	}

	url := fmt.Sprintf("%s/get-my-status", s.cfg.SMSBaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("send status request: %w", err)
	}
	defer resp.Body.Close()

	var statusResp SMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return "", fmt.Errorf("decode status response: %w", err)
	}

	return statusResp.QueryState, nil
}
