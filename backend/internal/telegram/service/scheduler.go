package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/cakeru/autostock/internal/telegram/models"
)

func gzipBytes(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var phnomPenh = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Phnom_Penh")
	if err != nil {
		return time.FixedZone("ICT", 7*60*60)
	}
	return loc
}()

// RunScheduler checks, roughly every 15 minutes, whether any time-based
// digest is due for any branch. Due digests are enqueued as telegram_events
// so the same delivery loop that handles real-time notifications delivers
// them too.
func (s *Service) RunScheduler(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	s.checkSchedule(ctx) // catch up once immediately on boot
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkSchedule(ctx)
		}
	}
}

func (s *Service) checkSchedule(ctx context.Context) {
	rows, err := s.pool.Query(ctx, `SELECT id FROM branches`)
	if err != nil {
		log.Printf("telegram: list branches: %v", err)
		return
	}
	var branchIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			branchIDs = append(branchIDs, id)
		}
	}
	rows.Close()

	now := time.Now().In(phnomPenh)
	today := now.Format("2006-01-02")

	for _, branchID := range branchIDs {
		last := s.scheduleMarkers(ctx, branchID)

		if now.Hour() >= 20 && last["daily_digest"] != today {
			if err := s.produceDailyDigest(ctx, branchID, today); err != nil {
				log.Printf("telegram: daily digest for branch %d: %v", branchID, err)
			}
			s.setMarker(ctx, branchID, "telegram_last_daily_digest", today)
		}

		if now.Hour() >= 7 && last["tomorrow_appts"] != today {
			if err := s.produceTomorrowAppts(ctx, branchID, now.AddDate(0, 0, 1).Format("2006-01-02")); err != nil {
				log.Printf("telegram: tomorrow appts for branch %d: %v", branchID, err)
			}
			s.setMarker(ctx, branchID, "telegram_last_tomorrow_appts", today)
		}

		// Once a week (Monday morning) rather than daily — the due-for-service
		// list barely moves day to day, and a daily nag would get ignored.
		if now.Weekday() == time.Monday && now.Hour() >= 9 && last["due_for_service"] != today {
			if err := s.produceDueForService(ctx, branchID); err != nil {
				log.Printf("telegram: due-for-service digest for branch %d: %v", branchID, err)
			}
			s.setMarker(ctx, branchID, "telegram_last_due_for_service", today)
		}

		weekStart := now.AddDate(0, 0, -int(now.Weekday())+1).Format("2006-01-02") // Monday of this week
		if now.Weekday() == time.Monday && now.Hour() >= 8 && last["weekly_ap"] != weekStart {
			if err := s.produceWeeklyAP(ctx, branchID); err != nil {
				log.Printf("telegram: weekly AP for branch %d: %v", branchID, err)
			}
			s.setMarker(ctx, branchID, "telegram_last_weekly_ap", weekStart)
		}

		monthMarker := now.Format("2006-01")
		if now.Day() == 1 && now.Hour() >= 6 {
			if last["monthly_report"] != monthMarker {
				if err := s.produceMonthlyReport(ctx, branchID, now); err != nil {
					log.Printf("telegram: monthly report for branch %d: %v", branchID, err)
				}
				s.setMarker(ctx, branchID, "telegram_last_monthly_report", monthMarker)
			}
			if last["monthly_backup"] != monthMarker {
				if err := s.produceMonthlyBackup(ctx, branchID); err != nil {
					log.Printf("telegram: monthly backup for branch %d: %v", branchID, err)
				}
				s.setMarker(ctx, branchID, "telegram_last_monthly_backup", monthMarker)
			}
		}
	}
}

func (s *Service) scheduleMarkers(ctx context.Context, branchID int64) map[string]string {
	out := map[string]string{}
	rows, err := s.pool.Query(ctx, `
		SELECT key, value FROM settings WHERE branch_id = $1 AND key IN
		('telegram_last_daily_digest', 'telegram_last_tomorrow_appts', 'telegram_last_weekly_ap', 'telegram_last_monthly_report', 'telegram_last_monthly_backup', 'telegram_last_due_for_service')`,
		branchID)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		if rows.Scan(&key, &value) == nil {
			out[strings.TrimPrefix(key, "telegram_last_")] = value
		}
	}
	return out
}

func (s *Service) setMarker(ctx context.Context, branchID int64, key, value string) {
	if err := s.upsertSetting(ctx, branchID, key, value); err != nil {
		log.Printf("telegram: set marker %s: %v", key, err)
	}
}

func (s *Service) insertEvent(ctx context.Context, branchID int64, topic, eventType string, payload map[string]any) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO telegram_events (branch_id, topic, event_type, payload) VALUES ($1, $2, $3, $4)`,
		branchID, topic, eventType, payloadJSON)
	return err
}

// TriggerNow enqueues a scheduled topic immediately, ignoring the normal
// once-a-day/week/month gating — for an admin testing their setup, or
// wanting an off-schedule report right now. Does not touch the schedule
// markers, so it doesn't affect when the topic would next fire automatically.
func (s *Service) TriggerNow(ctx context.Context, branchID int64, topic string) error {
	now := time.Now().In(phnomPenh)
	switch topic {
	case models.TopicDailyDigest:
		return s.produceDailyDigest(ctx, branchID, now.Format("2006-01-02"))
	case models.TopicTomorrowAppts:
		return s.produceTomorrowAppts(ctx, branchID, now.AddDate(0, 0, 1).Format("2006-01-02"))
	case models.TopicWeeklyAP:
		return s.produceWeeklyAP(ctx, branchID)
	case models.TopicMonthlyReport:
		return s.produceMonthlyReport(ctx, branchID, now)
	case models.TopicMonthlyBackup:
		return s.produceMonthlyBackup(ctx, branchID)
	case models.TopicDueForService:
		return s.produceDueForService(ctx, branchID)
	default:
		return fmt.Errorf("topic %q can't be triggered manually", topic)
	}
}

// ---------------------------------------------------------------------------
// Digest producers — compute the data and enqueue it; rendering happens at
// delivery time in renderScheduledDigest.
// ---------------------------------------------------------------------------

func (s *Service) produceDailyDigest(ctx context.Context, branchID int64, date string) error {
	var salesTotal, cashCollected float64
	var invoiceCount, jobsCompleted int

	_ = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_usd),0), COUNT(*) FROM invoices WHERE branch_id=$1 AND status<>'voided' AND issued_at::date = $2::date`,
		branchID, date).Scan(&salesTotal, &invoiceCount)
	_ = s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM service_jobs WHERE branch_id=$1 AND status='completed' AND completed_at::date = $2::date`,
		branchID, date).Scan(&jobsCompleted)
	_ = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(p.amount),0) FROM payments p JOIN invoices i ON i.id=p.invoice_id
		 WHERE i.branch_id=$1 AND p.created_at::date = $2::date`,
		branchID, date).Scan(&cashCollected)

	return s.insertEvent(ctx, branchID, models.TopicDailyDigest, "scheduled", map[string]any{
		"digest": "daily", "date": date,
		"sales_total": salesTotal, "invoice_count": invoiceCount,
		"jobs_completed": jobsCompleted, "cash_collected": cashCollected,
	})
}

func (s *Service) produceTomorrowAppts(ctx context.Context, branchID int64, date string) error {
	rows, err := s.pool.Query(ctx, `
		SELECT sj.job_number, sj.scheduled_at, COALESCE(c.name, 'Walk-in'), COALESCE(sj.description, '')
		FROM service_jobs sj LEFT JOIN customers c ON c.id = sj.customer_id
		WHERE sj.branch_id = $1 AND sj.scheduled_at::date = $2::date AND sj.status NOT IN ('completed', 'cancelled')
		ORDER BY sj.scheduled_at`, branchID, date)
	if err != nil {
		return fmt.Errorf("query tomorrow appts: %w", err)
	}
	defer rows.Close()

	var jobs []map[string]any
	for rows.Next() {
		var jobNumber, customerName, description string
		var scheduledAt time.Time
		if err := rows.Scan(&jobNumber, &scheduledAt, &customerName, &description); err != nil {
			continue
		}
		jobs = append(jobs, map[string]any{
			"job_number": jobNumber, "time": scheduledAt.In(phnomPenh).Format("15:04"),
			"customer_name": customerName, "description": description,
		})
	}

	return s.insertEvent(ctx, branchID, models.TopicTomorrowAppts, "scheduled", map[string]any{
		"digest": "tomorrow_appts", "date": date, "jobs": jobs,
	})
}

func (s *Service) produceWeeklyAP(ctx context.Context, branchID int64) error {
	rows, err := s.pool.Query(ctx, `
		SELECT sup.name, SUM(b.quantity_received * b.unit_cost - b.amount_paid) AS outstanding,
		       MAX(CURRENT_DATE - b.received_at::date) AS oldest_days
		FROM suppliers sup JOIN batches b ON b.supplier_id = sup.id
		WHERE sup.branch_id = $1 AND (b.quantity_received * b.unit_cost - b.amount_paid) > 0.01
		GROUP BY sup.name ORDER BY outstanding DESC`, branchID)
	if err != nil {
		return fmt.Errorf("query weekly AP: %w", err)
	}
	defer rows.Close()

	var suppliers []map[string]any
	for rows.Next() {
		var name string
		var outstanding float64
		var oldestDays int
		if err := rows.Scan(&name, &outstanding, &oldestDays); err != nil {
			continue
		}
		suppliers = append(suppliers, map[string]any{"name": name, "outstanding": outstanding, "oldest_days": oldestDays})
	}

	return s.insertEvent(ctx, branchID, models.TopicWeeklyAP, "scheduled", map[string]any{
		"digest": "weekly_ap", "suppliers": suppliers,
	})
}

func (s *Service) produceMonthlyReport(ctx context.Context, branchID int64, now time.Time) error {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, phnomPenh).AddDate(0, -1, 0)
	from := monthStart.Format("2006-01-02")
	to := monthStart.AddDate(0, 1, 0).AddDate(0, 0, -1).Format("2006-01-02")

	pnl, err := s.analytics.GetPnL(ctx, branchID, from, to)
	if err != nil {
		return fmt.Errorf("compute monthly PnL: %w", err)
	}

	return s.insertEvent(ctx, branchID, models.TopicMonthlyReport, "scheduled", map[string]any{
		"digest": "monthly_report", "from": from, "to": to,
		"revenue": pnl.Revenue, "cogs": pnl.COGS, "gross_profit": pnl.GrossProfit,
		"payroll": pnl.Payroll, "expenses": pnl.Expenses, "net_profit": pnl.NetProfit,
	})
}

func (s *Service) produceDueForService(ctx context.Context, branchID int64) error {
	items, err := s.vehicle.ListDueForService(ctx, branchID, 0)
	if err != nil {
		return fmt.Errorf("list due for service: %w", err)
	}

	var out []map[string]any
	for _, i := range items {
		dueDate := ""
		if i.DueDate != nil {
			dueDate = i.DueDate.Format("2006-01-02")
		}
		out = append(out, map[string]any{
			"plate_number":   i.PlateNumber,
			"customer_name":  i.CustomerName,
			"customer_phone": i.CustomerPhone,
			"event_type":     i.EventType,
			"label":          i.Label,
			"status":         i.Status,
			"due_date":       dueDate,
		})
	}

	return s.insertEvent(ctx, branchID, models.TopicDueForService, "scheduled", map[string]any{
		"digest": "due_for_service", "items": out,
	})
}

func (s *Service) produceMonthlyBackup(ctx context.Context, branchID int64) error {
	// The dump itself is taken fresh at delivery time (see deliverMonthlyBackup)
	// rather than precomputed here, so it reflects the database as of when it's
	// actually sent, not when it was scheduled.
	return s.insertEvent(ctx, branchID, models.TopicMonthlyBackup, "monthly_backup", map[string]any{})
}

func (s *Service) deliverMonthlyBackup(ctx context.Context, channel models.Channel, e pendingEvent) error {
	dsn := s.databaseURL
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL not configured for backup")
	}
	cmd := exec.CommandContext(ctx, "pg_dump", dsn, "--no-owner", "--no-privileges")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failed: %w: %s", err, stderr.String())
	}

	gzipped, err := gzipBytes([]byte(stdout.String()))
	if err != nil {
		return fmt.Errorf("gzip backup: %w", err)
	}

	filename := fmt.Sprintf("autostock-backup-%s.sql.gz", time.Now().In(phnomPenh).Format("2006-01-02"))
	caption := fmt.Sprintf("📦 Monthly database backup — %s", time.Now().In(phnomPenh).Format("2 Jan 2006"))
	return s.sendDocument(ctx, channel.BotToken, channel.ChatID, filename, gzipped, caption)
}

// renderScheduledDigest formats the "scheduled" event_type based on which
// digest producer created it (the "digest" discriminator field).
func renderScheduledDigest(digest string, p map[string]any) string {
	str := func(k string) string { v, _ := p[k].(string); return v }
	num := func(k string) float64 { v, _ := p[k].(float64); return v }

	switch digest {
	case "daily":
		return fmt.Sprintf("📊 <b>Daily summary — %s</b>\nSales: $%.2f (%d invoice(s))\nJobs completed: %d\nCash collected: $%.2f",
			str("date"), num("sales_total"), int(num("invoice_count")), int(num("jobs_completed")), num("cash_collected"))
	case "tomorrow_appts":
		jobsRaw, _ := p["jobs"].([]any)
		if len(jobsRaw) == 0 {
			return fmt.Sprintf("🗓 <b>Tomorrow (%s)</b>\nNo appointments scheduled.", str("date"))
		}
		var lines []string
		for _, j := range jobsRaw {
			jm, _ := j.(map[string]any)
			jt, _ := jm["time"].(string)
			jn, _ := jm["job_number"].(string)
			cn, _ := jm["customer_name"].(string)
			desc, _ := jm["description"].(string)
			lines = append(lines, fmt.Sprintf("• %s — %s (%s) %s", jt, jn, cn, desc))
		}
		return fmt.Sprintf("🗓 <b>Tomorrow (%s)</b>\n%s", str("date"), strings.Join(lines, "\n"))
	case "weekly_ap":
		suppliersRaw, _ := p["suppliers"].([]any)
		if len(suppliersRaw) == 0 {
			return "💳 <b>Supplier payables</b>\nNothing outstanding."
		}
		var lines []string
		for _, sup := range suppliersRaw {
			sm, _ := sup.(map[string]any)
			name, _ := sm["name"].(string)
			outstanding, _ := sm["outstanding"].(float64)
			oldest, _ := sm["oldest_days"].(float64)
			lines = append(lines, fmt.Sprintf("• %s — $%.2f (oldest %d day(s))", name, outstanding, int(oldest)))
		}
		return fmt.Sprintf("💳 <b>Supplier payables outstanding</b>\n%s", strings.Join(lines, "\n"))
	case "monthly_report":
		return fmt.Sprintf("📈 <b>Monthly P&amp;L — %s to %s</b>\nRevenue: $%.2f\nCOGS: $%.2f\nGross profit: $%.2f\nPayroll: $%.2f\nExpenses: $%.2f\n<b>Net profit: $%.2f</b>",
			str("from"), str("to"), num("revenue"), num("cogs"), num("gross_profit"), num("payroll"), num("expenses"), num("net_profit"))
	case "due_for_service":
		itemsRaw, _ := p["items"].([]any)
		if len(itemsRaw) == 0 {
			return "🔔 <b>Due for service</b>\nNobody's overdue or coming up this week."
		}
		var overdue, dueSoon []string
		for _, it := range itemsRaw {
			im, _ := it.(map[string]any)
			plate, _ := im["plate_number"].(string)
			customer, _ := im["customer_name"].(string)
			phone, _ := im["customer_phone"].(string)
			eventType, _ := im["event_type"].(string)
			status, _ := im["status"].(string)
			due, _ := im["due_date"].(string)
			kind, _ := im["label"].(string)
			if kind == "" { // older payloads had no label
				kind = "Oil"
				if eventType == "tire" {
					kind = "Tires"
				}
			}
			phoneSuffix := ""
			if phone != "" {
				phoneSuffix = " · " + phone
			}
			line := fmt.Sprintf("• %s — %s (%s)%s, due %s", plate, customer, kind, phoneSuffix, orDash(due))
			if status == "overdue" {
				overdue = append(overdue, line)
			} else {
				dueSoon = append(dueSoon, line)
			}
		}
		var sections []string
		if len(overdue) > 0 {
			sections = append(sections, "<b>Overdue:</b>\n"+strings.Join(overdue, "\n"))
		}
		if len(dueSoon) > 0 {
			sections = append(sections, "<b>Due soon:</b>\n"+strings.Join(dueSoon, "\n"))
		}
		return "🔔 <b>Due for service</b>\n" + strings.Join(sections, "\n\n")
	default:
		return "AutoStock scheduled update"
	}
}
