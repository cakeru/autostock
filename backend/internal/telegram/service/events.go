package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/cakeru/autostock/internal/telegram/models"
)

// LogEvent records a notification-worthy thing that just happened, inside the
// same transaction as the change itself — so a notification only ever exists
// for a change that actually committed. A background loop (RunDeliveryLoop)
// picks it up and delivers it independently, so a slow or unreachable
// Telegram API never adds latency or risk to the business operation that
// triggered it. Mirrors the batch package's style: a plain function taking a
// tx, no service instance required to log.
func LogEvent(ctx context.Context, tx pgx.Tx, branchID int64, topic, eventType, refType string, refID int64, payload map[string]any) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal telegram event payload: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO telegram_events (branch_id, topic, event_type, reference_type, reference_id, payload)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		branchID, topic, eventType, refType, refID, payloadJSON)
	if err != nil {
		return fmt.Errorf("log telegram event: %w", err)
	}
	return nil
}

type pendingEvent struct {
	id          int64
	branchID    int64
	topic       string
	eventType   string
	referenceID *int64
	payload     map[string]any
}

// RunDeliveryLoop polls for pending events and attempts to deliver each one.
// Meant to run for the process lifetime as its own goroutine.
func (s *Service) RunDeliveryLoop(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processPendingEvents(ctx)
		}
	}
}

func (s *Service) processPendingEvents(ctx context.Context) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, branch_id, topic, event_type, reference_id, payload
		FROM telegram_events WHERE status = 'pending' ORDER BY id LIMIT 20`)
	if err != nil {
		log.Printf("telegram: query pending events: %v", err)
		return
	}
	var events []pendingEvent
	for rows.Next() {
		var e pendingEvent
		var payloadJSON []byte
		if err := rows.Scan(&e.id, &e.branchID, &e.topic, &e.eventType, &e.referenceID, &payloadJSON); err != nil {
			log.Printf("telegram: scan pending event: %v", err)
			continue
		}
		_ = json.Unmarshal(payloadJSON, &e.payload)
		events = append(events, e)
	}
	rows.Close()

	for _, e := range events {
		s.deliver(ctx, e)
	}
}

func (s *Service) deliver(ctx context.Context, e pendingEvent) {
	cfg, err := s.loadConfig(ctx, e.branchID)
	if err != nil {
		s.markEvent(ctx, e.id, "failed", err.Error())
		return
	}
	channel, ok := cfg.RouteChannel(e.topic)
	if !ok {
		s.markEvent(ctx, e.id, "skipped", "topic not routed to a channel")
		return
	}

	var deliverErr error
	switch e.eventType {
	case "job_created":
		deliverErr = s.deliverJobCreated(ctx, channel, e)
	case "job_status_changed":
		deliverErr = s.deliverJobStatusChanged(ctx, channel, e)
	case "monthly_backup":
		deliverErr = s.deliverMonthlyBackup(ctx, channel, e)
	default:
		text := renderMessage(e.eventType, e.payload)
		_, deliverErr = s.sendMessage(ctx, channel.BotToken, channel.ChatID, text)
	}

	if deliverErr != nil {
		s.markEvent(ctx, e.id, "failed", deliverErr.Error())
		return
	}
	s.markEvent(ctx, e.id, "sent", "")
}

func (s *Service) deliverJobCreated(ctx context.Context, channel models.Channel, e pendingEvent) error {
	text := renderMessage(e.eventType, e.payload)
	msgID, err := s.sendMessage(ctx, channel.BotToken, channel.ChatID, text)
	if err != nil {
		return err
	}
	if e.referenceID != nil {
		if _, err := s.pool.Exec(ctx,
			`UPDATE service_jobs SET telegram_chat_id = $1, telegram_message_id = $2 WHERE id = $3`,
			channel.ChatID, msgID, *e.referenceID); err != nil {
			log.Printf("telegram: link job message: %v", err)
		}
	}
	return nil
}

func (s *Service) deliverJobStatusChanged(ctx context.Context, channel models.Channel, e pendingEvent) error {
	text := renderMessage(e.eventType, e.payload)
	if e.referenceID == nil {
		_, err := s.sendMessage(ctx, channel.BotToken, channel.ChatID, text)
		return err
	}
	var chatID string
	var msgID *int64
	err := s.pool.QueryRow(ctx,
		`SELECT telegram_chat_id, telegram_message_id FROM service_jobs WHERE id = $1`, *e.referenceID).
		Scan(&chatID, &msgID)
	if err != nil || msgID == nil || chatID == "" {
		// No prior message to edit (e.g. its own job_created delivery hasn't
		// landed yet, or failed) — fall back to a fresh message.
		_, sendErr := s.sendMessage(ctx, channel.BotToken, channel.ChatID, text)
		return sendErr
	}
	return s.editMessageText(ctx, channel.BotToken, chatID, *msgID, text)
}

func (s *Service) markEvent(ctx context.Context, id int64, status, errMsg string) {
	_, err := s.pool.Exec(ctx,
		`UPDATE telegram_events SET status = $1::varchar, error = NULLIF($2, ''), sent_at = CASE WHEN $1::varchar = 'sent' THEN NOW() ELSE sent_at END WHERE id = $3`,
		status, errMsg, id)
	if err != nil {
		log.Printf("telegram: mark event %d as %s: %v", id, status, err)
	}
}

// ---------------------------------------------------------------------------
// Message rendering — one case per event_type, reading whatever fields that
// event_type's producer put in the payload.
// ---------------------------------------------------------------------------

func renderMessage(eventType string, p map[string]any) string {
	str := func(k string) string { v, _ := p[k].(string); return v }
	num := func(k string) float64 { v, _ := p[k].(float64); return v }

	switch eventType {
	case "job_created":
		return fmt.Sprintf("🔧 <b>New job %s</b>\nCustomer: %s%s\n%s\nAssigned: %s\nStatus: %s",
			str("job_number"), orWalkin(str("customer_name")),
			vehicleSuffix(str("vehicle_info")), orDash(str("description")),
			orDash(str("assigned_to_name")), str("status"))
	case "job_status_changed":
		return fmt.Sprintf("🔧 <b>Job %s</b>\nCustomer: %s%s\n%s\nAssigned: %s\nStatus: %s → <b>%s</b>",
			str("job_number"), orWalkin(str("customer_name")),
			vehicleSuffix(str("vehicle_info")), orDash(str("description")),
			orDash(str("assigned_to_name")), str("old_status"), str("new_status"))
	case "sale_issued":
		return fmt.Sprintf("💰 <b>Sale %s</b>\nCustomer: %s\nTotal: $%.2f (%d item(s))",
			str("invoice_number"), orWalkin(str("customer_name")), num("total_usd"), int(num("item_count")))
	case "sale_voided":
		return fmt.Sprintf("⚠️ <b>Invoice %s voided</b>\nReason: %s",
			str("invoice_number"), orDash(str("reason")))
	case "cash_discrepancy":
		return fmt.Sprintf("🚨 <b>Cash drawer discrepancy</b>\nExpected: $%.2f\nActual: $%.2f\nOver/short: $%.2f",
			num("expected"), num("actual"), num("over_short"))
	case "low_stock":
		return fmt.Sprintf("📉 <b>Low stock</b>\n%s (%s)\nOn hand: %d — below alert level of %d",
			str("name"), str("sku"), int(num("stock_qty")), int(num("min_alert")))
	case "stocktake_shrinkage":
		return fmt.Sprintf("📦 <b>Stocktake shrinkage</b>\n%s (%s)\nVariance: %d units",
			str("name"), str("sku"), int(num("variance")))
	case "scheduled":
		return renderScheduledDigest(str("digest"), p)
	default:
		return fmt.Sprintf("AutoStock notification (%s)", eventType)
	}
}

func orWalkin(s string) string {
	if s == "" {
		return "Walk-in"
	}
	return s
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func vehicleSuffix(v string) string {
	if v == "" {
		return ""
	}
	return " — " + v
}
