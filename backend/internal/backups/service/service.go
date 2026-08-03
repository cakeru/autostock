package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"

	"github.com/cakeru/autostock/internal/backups/dto"
	"github.com/cakeru/autostock/internal/domain"
)

type Schedule struct {
	ID            int64
	Name          string
	Cron          string
	Enabled       bool
	RetentionDays int
	LastRunAt     *time.Time
	LastStatus    string
	LastError     string
	NextRunAt     *time.Time
}

type Service struct {
	pool      *pgxpool.Pool
	backupDir string
	dbURL     string

	// dumpMu serializes pg_dump so overlapping schedules never run concurrently.
	dumpMu sync.Mutex
}

func NewService(pool *pgxpool.Pool, backupDir, dbURL string) *Service {
	return &Service{pool: pool, backupDir: backupDir, dbURL: dbURL}
}

func validateCron(expr string) error {
	if _, err := cron.ParseStandard(expr); err != nil {
		return &domain.AppError{Code: "INVALID_CRON", Message: "Invalid schedule expression: " + err.Error(), Status: 400}
	}
	return nil
}

func nextRun(expr string, from time.Time) *time.Time {
	sched, err := cron.ParseStandard(expr)
	if err != nil {
		return nil
	}
	n := sched.Next(from)
	return &n
}

func (s *Service) List(ctx context.Context) ([]dto.ScheduleResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, cron, enabled, retention_days, last_run_at, last_status, last_error, created_at
		FROM backup_schedules ORDER BY enabled DESC, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	var out []dto.ScheduleResponse
	for rows.Next() {
		var sch dto.ScheduleResponse
		if err := rows.Scan(&sch.ID, &sch.Name, &sch.Cron, &sch.Enabled, &sch.RetentionDays,
			&sch.LastRunAt, &sch.LastStatus, &sch.LastError, &sch.CreatedAt); err != nil {
			return nil, err
		}
		if sch.Enabled {
			sch.NextRunAt = nextRun(sch.Cron, now)
		}
		sch.LatestFile = s.latestFileFor(sch.ID)
		out = append(out, sch)
	}
	return out, rows.Err()
}

func (s *Service) Create(ctx context.Context, req dto.ScheduleRequest) (*dto.ScheduleResponse, error) {
	if err := validateCron(req.Cron); err != nil {
		return nil, err
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	var sch dto.ScheduleResponse
	err := s.pool.QueryRow(ctx, `
		INSERT INTO backup_schedules (name, cron, enabled, retention_days)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, cron, enabled, retention_days, last_run_at, last_status, last_error, created_at`,
		strings.TrimSpace(req.Name), req.Cron, enabled, req.RetentionDays).
		Scan(&sch.ID, &sch.Name, &sch.Cron, &sch.Enabled, &sch.RetentionDays,
			&sch.LastRunAt, &sch.LastStatus, &sch.LastError, &sch.CreatedAt)
	if err != nil {
		return nil, err
	}
	if sch.Enabled {
		sch.NextRunAt = nextRun(sch.Cron, time.Now())
	}
	return &sch, nil
}

func (s *Service) Update(ctx context.Context, id int64, req dto.ScheduleRequest) (*dto.ScheduleResponse, error) {
	if err := validateCron(req.Cron); err != nil {
		return nil, err
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	var sch dto.ScheduleResponse
	err := s.pool.QueryRow(ctx, `
		UPDATE backup_schedules
		SET name = $1, cron = $2, enabled = $3, retention_days = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING id, name, cron, enabled, retention_days, last_run_at, last_status, last_error, created_at`,
		strings.TrimSpace(req.Name), req.Cron, enabled, req.RetentionDays, id).
		Scan(&sch.ID, &sch.Name, &sch.Cron, &sch.Enabled, &sch.RetentionDays,
			&sch.LastRunAt, &sch.LastStatus, &sch.LastError, &sch.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if sch.Enabled {
		sch.NextRunAt = nextRun(sch.Cron, time.Now())
	}
	return &sch, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM backup_schedules WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Service) RunNow(ctx context.Context, id int64) (*dto.ScheduleResponse, error) {
	sch, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.runOne(ctx, sch); err != nil {
		return nil, err
	}
	updated, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	res := toResponse(updated)
	if res.Enabled {
		res.NextRunAt = nextRun(res.Cron, time.Now())
	}
	res.LatestFile = s.latestFileFor(res.ID)
	return &res, nil
}

func (s *Service) get(ctx context.Context, id int64) (*Schedule, error) {
	var sch Schedule
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, cron, enabled, retention_days, last_run_at, last_status, last_error
		FROM backup_schedules WHERE id = $1`, id).
		Scan(&sch.ID, &sch.Name, &sch.Cron, &sch.Enabled, &sch.RetentionDays,
			&sch.LastRunAt, &sch.LastStatus, &sch.LastError)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sch, nil
}

func toResponse(sch *Schedule) dto.ScheduleResponse {
	return dto.ScheduleResponse{
		ID:            sch.ID,
		Name:          sch.Name,
		Cron:          sch.Cron,
		Enabled:       sch.Enabled,
		RetentionDays: sch.RetentionDays,
		LastRunAt:     sch.LastRunAt,
		LastStatus:    sch.LastStatus,
		LastError:     sch.LastError,
	}
}

// Run is the scheduler loop: every 30s it fires any enabled schedule whose
// next_run_at is due (a NULL next_run_at runs on boot, so a fresh install gets
// a backup quickly rather than waiting a day).
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		s.runDue(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) runDue(ctx context.Context) {
	rows, err := s.pool.Query(ctx, `
		SELECT id FROM backup_schedules
		WHERE enabled AND (next_run_at IS NULL OR next_run_at <= NOW())`)
	if err != nil {
		return
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		sch, err := s.get(ctx, id)
		if err != nil {
			continue
		}
		go s.runOne(context.WithoutCancel(ctx), sch)
	}
}

func (s *Service) runOne(ctx context.Context, sch *Schedule) error {
	s.dumpMu.Lock()
	defer s.dumpMu.Unlock()

	_, err := s.dumpToFile(ctx, sch)
	status, errMsg := "success", ""
	if err != nil {
		status, errMsg = "error", err.Error()
	}
	next := nextRun(sch.Cron, time.Now())
	_, _ = s.pool.Exec(ctx, `
		UPDATE backup_schedules
		SET last_run_at = NOW(), last_status = $1, last_error = $2, next_run_at = $3, updated_at = NOW()
		WHERE id = $4`, status, errMsg, next, sch.ID)
	return err
}

func (s *Service) dumpToFile(ctx context.Context, sch *Schedule) (string, error) {
	if s.dbURL == "" {
		return "", fmt.Errorf("DATABASE_URL not configured for backups")
	}
	if err := os.MkdirAll(s.backupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}

	cmd := exec.CommandContext(ctx, "pg_dump", s.dbURL, "--no-owner", "--no-privileges")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pg_dump failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	if _, err := zw.Write(stdout.Bytes()); err != nil {
		return "", fmt.Errorf("gzip: %w", err)
	}
	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("gzip close: %w", err)
	}

	slug := slugify(sch.Name)
	filename := fmt.Sprintf("autostock-%d-%s-%s.sql.gz", sch.ID, slug, time.Now().Format("2006-01-02-150405"))
	if err := os.WriteFile(filepath.Join(s.backupDir, filename), gz.Bytes(), 0o644); err != nil {
		return "", fmt.Errorf("write dump: %w", err)
	}
	s.cleanupOld(sch)
	return filename, nil
}

// cleanupOld deletes this schedule's dumps older than retention_days, matching
// the old compose job's "keep 14 days" behaviour.
func (s *Service) cleanupOld(sch *Schedule) {
	prefix := fmt.Sprintf("autostock-%d-", sch.ID)
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -sch.RetentionDays)
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), prefix) || !strings.HasSuffix(e.Name(), ".sql.gz") {
			continue
		}
		info, err := e.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(s.backupDir, e.Name()))
		}
	}
}

// latestFileFor returns the newest dump filename for a schedule, if any.
func (s *Service) latestFileFor(id int64) string {
	prefix := fmt.Sprintf("autostock-%d-", id)
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return ""
	}
	var best string
	var bestTime time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), prefix) || !strings.HasSuffix(e.Name(), ".sql.gz") {
			continue
		}
		if info, err := e.Info(); err == nil && info.ModTime().After(bestTime) {
			bestTime = info.ModTime()
			best = e.Name()
		}
	}
	return best
}

// LatestFile returns the on-disk path of a schedule's newest dump, or ("", nil)
// when the schedule exists but has never produced one.
func (s *Service) LatestFile(ctx context.Context, id int64) (string, error) {
	if _, err := s.get(ctx, id); err != nil {
		return "", err
	}
	filename := s.latestFileFor(id)
	if filename == "" {
		return "", nil
	}
	return filepath.Join(s.backupDir, filename), nil
}

var slugRe = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func slugify(name string) string {
	slug := slugRe.ReplaceAllString(name, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "backup"
	}
	if len(slug) > 40 {
		slug = slug[:40]
	}
	return strings.ToLower(slug)
}
