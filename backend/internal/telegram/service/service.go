package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"

	analyticsService "github.com/cakeru/autostock/internal/analytics/service"
	"github.com/cakeru/autostock/internal/telegram/dto"
	"github.com/cakeru/autostock/internal/telegram/models"
	vehicleService "github.com/cakeru/autostock/internal/vehicle/service"
)

type Service struct {
	pool        *pgxpool.Pool
	analytics   *analyticsService.Service
	vehicle     *vehicleService.Service
	http        *http.Client
	databaseURL string
}

func NewService(pool *pgxpool.Pool, analytics *analyticsService.Service, vehicle *vehicleService.Service, databaseURL string) *Service {
	return &Service{pool: pool, analytics: analytics, vehicle: vehicle, databaseURL: databaseURL, http: &http.Client{Timeout: 15 * time.Second}}
}

// ---------------------------------------------------------------------------
// Config: channels + topic routing, stored as JSON in the generic settings
// table (same pattern already used for labor/fee presets) — no dedicated
// tables needed since this is an admin-managed, wholesale-replaced list.
// ---------------------------------------------------------------------------

func (s *Service) loadConfig(ctx context.Context, branchID int64) (*models.Config, error) {
	cfg := &models.Config{Routes: map[string]string{}}

	var channelsJSON, routesJSON string
	rows, err := s.pool.Query(ctx,
		`SELECT key, value FROM settings WHERE branch_id = $1 AND key IN ('telegram_channels', 'telegram_routes')`,
		branchID)
	if err != nil {
		return nil, fmt.Errorf("load telegram config: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		switch key {
		case "telegram_channels":
			channelsJSON = value
		case "telegram_routes":
			routesJSON = value
		}
	}

	if channelsJSON != "" {
		if err := json.Unmarshal([]byte(channelsJSON), &cfg.Channels); err != nil {
			return nil, fmt.Errorf("parse telegram_channels: %w", err)
		}
	}
	if routesJSON != "" {
		if err := json.Unmarshal([]byte(routesJSON), &cfg.Routes); err != nil {
			return nil, fmt.Errorf("parse telegram_routes: %w", err)
		}
	}
	return cfg, nil
}

func (s *Service) GetChannels(ctx context.Context, branchID int64) ([]dto.Channel, error) {
	cfg, err := s.loadConfig(ctx, branchID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.Channel, 0, len(cfg.Channels))
	for _, ch := range cfg.Channels {
		out = append(out, dto.Channel(ch))
	}
	return out, nil
}

func (s *Service) SaveChannels(ctx context.Context, branchID int64, channels []dto.Channel) error {
	payload, err := json.Marshal(channels)
	if err != nil {
		return fmt.Errorf("marshal channels: %w", err)
	}
	return s.upsertSetting(ctx, branchID, "telegram_channels", string(payload))
}

func (s *Service) GetRoutes(ctx context.Context, branchID int64) (map[string]string, error) {
	cfg, err := s.loadConfig(ctx, branchID)
	if err != nil {
		return nil, err
	}
	return cfg.Routes, nil
}

func (s *Service) SaveRoutes(ctx context.Context, branchID int64, routes map[string]string) error {
	payload, err := json.Marshal(routes)
	if err != nil {
		return fmt.Errorf("marshal routes: %w", err)
	}
	return s.upsertSetting(ctx, branchID, "telegram_routes", string(payload))
}

func (s *Service) upsertSetting(ctx context.Context, branchID int64, key, value string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO settings (branch_id, key, value) VALUES ($1, $2, $3)
		 ON CONFLICT (branch_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
		branchID, key, value)
	return err
}

// TestSend sends a short confirmation message to one configured channel, so
// an admin can verify a bot token + chat ID actually work before relying on
// it for real notifications.
func (s *Service) TestSend(ctx context.Context, branchID int64, channelID string) error {
	cfg, err := s.loadConfig(ctx, branchID)
	if err != nil {
		return err
	}
	channel, ok := cfg.ChannelByID(channelID)
	if !ok {
		return fmt.Errorf("channel not found")
	}
	_, err = s.sendMessage(ctx, channel.BotToken, channel.ChatID, "✅ AutoStock test message — this channel is wired up correctly.")
	return err
}

// SendDocumentToTopic delivers an on-demand file (an invoice or vehicle-report
// PDF) to whichever channel the branch has routed the given topic to. Used by
// the "Send to Telegram" buttons so an admin can forward it to a customer.
// Returns a user-actionable error when the topic isn't routed yet.
func (s *Service) SendDocumentToTopic(ctx context.Context, branchID int64, topic, filename string, data []byte, caption string) error {
	cfg, err := s.loadConfig(ctx, branchID)
	if err != nil {
		return err
	}
	channel, ok := cfg.RouteChannel(topic)
	if !ok {
		return &domain.AppError{
			Code:    "TOPIC_NOT_ROUTED",
			Message: "No Telegram channel is set for documents yet. Set one under Settings → Telegram (route the \"Documents\" topic to a channel).",
			Status:  400,
		}
	}
	return s.sendDocument(ctx, channel.BotToken, channel.ChatID, filename, data, caption)
}

// ---------------------------------------------------------------------------
// Telegram Bot API client — plain HTTP, no SDK dependency.
// ---------------------------------------------------------------------------

type apiResponse struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

func (s *Service) apiPost(ctx context.Context, botToken, method string, body map[string]any) (json.RawMessage, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	var out apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode telegram response: %w", err)
	}
	if !out.OK {
		return nil, fmt.Errorf("telegram API error: %s", out.Description)
	}
	return out.Result, nil
}

func (s *Service) sendMessage(ctx context.Context, botToken, chatID, text string) (int64, error) {
	result, err := s.apiPost(ctx, botToken, "sendMessage", map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	})
	if err != nil {
		return 0, err
	}
	var msg struct {
		MessageID int64 `json:"message_id"`
	}
	if err := json.Unmarshal(result, &msg); err != nil {
		return 0, fmt.Errorf("parse sendMessage result: %w", err)
	}
	return msg.MessageID, nil
}

func (s *Service) editMessageText(ctx context.Context, botToken, chatID string, messageID int64, text string) error {
	_, err := s.apiPost(ctx, botToken, "editMessageText", map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "HTML",
	})
	return err
}

func (s *Service) sendDocument(ctx context.Context, botToken, chatID, filename string, data []byte, caption string) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("chat_id", chatID)
	if caption != "" {
		_ = w.WriteField("caption", caption)
	}
	part, err := w.CreateFormFile("document", filename)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("write document: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	var out apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("decode telegram response: %w", err)
	}
	if !out.OK {
		return fmt.Errorf("telegram API error: %s", out.Description)
	}
	return nil
}
