package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/vehicle/dto"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// GetProfile returns the vehicle's specs, owner, and a lightweight summary
// (record count, most recent known mileage/service date across records,
// invoices, and jobs — whichever is freshest).
func (s *Service) GetProfile(ctx context.Context, branchID, id int64) (*dto.VehicleProfileResponse, error) {
	var v dto.VehicleProfileResponse
	var year *int
	err := s.pool.QueryRow(ctx, `
		SELECT vh.id, vh.plate_number, COALESCE(vh.make,''), COALESCE(vh.model,''), vh.year,
		       COALESCE(vh.vin,''), COALESCE(vh.color,''), COALESCE(vh.body_type,''), COALESCE(vh.notes,''),
		       vh.customer_id, COALESCE(c.name,''), COALESCE(c.phone,''), vh.created_at,
		       vh.oil_interval_km, vh.oil_interval_days, vh.tire_interval_km, vh.tire_interval_days,
		       vh.distance_unit, COALESCE(vh.share_token,'')
		FROM vehicles vh
		JOIN customers c ON c.id = vh.customer_id
		WHERE vh.id = $1 AND vh.branch_id = $2`, id, branchID).
		Scan(&v.ID, &v.PlateNumber, &v.Make, &v.Model, &year, &v.VIN, &v.Color, &v.BodyType, &v.Notes,
			&v.CustomerID, &v.CustomerName, &v.CustomerPhone, &v.CreatedAt,
			&v.OilIntervalKm, &v.OilIntervalDays, &v.TireIntervalKm, &v.TireIntervalDays, &v.DistanceUnit,
			&v.ShareToken)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get vehicle profile: %w", err)
	}
	if year != nil {
		v.Year = *year
	}

	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM vehicle_records WHERE vehicle_id = $1`, id).Scan(&v.RecordCount)

	_ = s.pool.QueryRow(ctx, `
		SELECT mileage, at FROM (
			SELECT mileage, created_at AS at FROM vehicle_records WHERE vehicle_id = $1 AND mileage IS NOT NULL
			UNION ALL
			SELECT mileage, COALESCE(issued_at, created_at) AS at FROM invoices WHERE vehicle_id = $1 AND mileage IS NOT NULL AND status <> 'voided'
			UNION ALL
			SELECT mileage, COALESCE(completed_at, created_at) AS at FROM service_jobs WHERE vehicle_id = $1 AND mileage IS NOT NULL
		) t ORDER BY at DESC LIMIT 1`, id).Scan(&v.LastMileage, &v.LastServiceAt)

	due, err := s.GetDueForVehicle(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	v.Due = due

	return &v, nil
}

// GetHistory returns this vehicle's invoice + job activity, newest first —
// the same shape the customer profile's activity timeline uses.
func (s *Service) GetHistory(ctx context.Context, branchID, vehicleID int64) ([]dto.HistoryItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT * FROM (
			SELECT 'job' AS type, sj.id, sj.job_number AS ref,
			       COALESCE(sj.completed_at, sj.created_at) AS date,
			       COALESCE(NULLIF(sj.description, ''), 'Service job') AS title,
			       sj.status,
			       COALESCE(SUM(sji.quantity * sji.unit_price), 0)::numeric(10,2) AS amount,
			       sj.mileage
			FROM service_jobs sj
			LEFT JOIN service_job_items sji ON sji.service_job_id = sj.id
			WHERE sj.vehicle_id = $1 AND sj.branch_id = $2
			GROUP BY sj.id
			UNION ALL
			SELECT 'invoice', i.id, i.invoice_number,
			       COALESCE(i.issued_at, i.created_at),
			       'Sale', i.status,
			       i.total_usd,
			       i.mileage
			FROM invoices i
			WHERE i.vehicle_id = $1 AND i.branch_id = $2 AND i.status <> 'voided'
		) t
		ORDER BY date DESC NULLS LAST
		LIMIT 50`, vehicleID, branchID)
	if err != nil {
		return nil, fmt.Errorf("query vehicle history: %w", err)
	}
	defer rows.Close()

	items := []dto.HistoryItem{}
	for rows.Next() {
		var it dto.HistoryItem
		if err := rows.Scan(&it.Type, &it.ID, &it.Ref, &it.Date, &it.Title, &it.Status, &it.Amount, &it.Mileage); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *Service) ListRecords(ctx context.Context, branchID, vehicleID int64) ([]dto.RecordResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.id, COALESCE(r.note,''), r.mileage, r.invoice_id, COALESCE(i.invoice_number,''),
		       r.service_job_id, COALESCE(sj.job_number,''), COALESCE(u.full_name,''), r.created_at
		FROM vehicle_records r
		LEFT JOIN invoices i ON i.id = r.invoice_id
		LEFT JOIN service_jobs sj ON sj.id = r.service_job_id
		LEFT JOIN users u ON u.id = r.created_by
		WHERE r.vehicle_id = $1 AND r.branch_id = $2
		ORDER BY r.created_at DESC`, vehicleID, branchID)
	if err != nil {
		return nil, fmt.Errorf("list records: %w", err)
	}

	var records []dto.RecordResponse
	var ids []int64
	for rows.Next() {
		var r dto.RecordResponse
		if err := rows.Scan(&r.ID, &r.Note, &r.Mileage, &r.InvoiceID, &r.InvoiceNumber,
			&r.ServiceJobID, &r.JobNumber, &r.CreatedByName, &r.CreatedAt); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan record: %w", err)
		}
		r.Photos = []dto.PhotoResponse{}
		records = append(records, r)
		ids = append(ids, r.ID)
	}
	rows.Close()

	if len(ids) > 0 {
		photoRows, err := s.pool.Query(ctx,
			`SELECT id, record_id, url FROM vehicle_record_photos WHERE record_id = ANY($1) ORDER BY created_at`, ids)
		if err != nil {
			return nil, fmt.Errorf("list photos: %w", err)
		}
		byRecord := map[int64][]dto.PhotoResponse{}
		for photoRows.Next() {
			var p dto.PhotoResponse
			var recordID int64
			if err := photoRows.Scan(&p.ID, &recordID, &p.URL); err != nil {
				photoRows.Close()
				return nil, fmt.Errorf("scan photo: %w", err)
			}
			byRecord[recordID] = append(byRecord[recordID], p)
		}
		photoRows.Close()
		for i := range records {
			if photos, ok := byRecord[records[i].ID]; ok {
				records[i].Photos = photos
			}
		}
	}

	if records == nil {
		records = []dto.RecordResponse{}
	}
	return records, nil
}

func (s *Service) getRecord(ctx context.Context, branchID, id int64) (*dto.RecordResponse, error) {
	var r dto.RecordResponse
	err := s.pool.QueryRow(ctx, `
		SELECT r.id, COALESCE(r.note,''), r.mileage, r.invoice_id, COALESCE(i.invoice_number,''),
		       r.service_job_id, COALESCE(sj.job_number,''), COALESCE(u.full_name,''), r.created_at
		FROM vehicle_records r
		LEFT JOIN invoices i ON i.id = r.invoice_id
		LEFT JOIN service_jobs sj ON sj.id = r.service_job_id
		LEFT JOIN users u ON u.id = r.created_by
		WHERE r.id = $1 AND r.branch_id = $2`, id, branchID).
		Scan(&r.ID, &r.Note, &r.Mileage, &r.InvoiceID, &r.InvoiceNumber,
			&r.ServiceJobID, &r.JobNumber, &r.CreatedByName, &r.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get record: %w", err)
	}
	r.Photos = []dto.PhotoResponse{}
	return &r, nil
}

func (s *Service) CreateRecord(ctx context.Context, branchID, vehicleID, userID int64, req *dto.CreateRecordRequest) (*dto.RecordResponse, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND branch_id = $2)`,
		vehicleID, branchID).Scan(&exists); err != nil || !exists {
		return nil, domain.ErrNotFound
	}

	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO vehicle_records (branch_id, vehicle_id, invoice_id, service_job_id, mileage, note, created_by)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7)
		RETURNING id`,
		branchID, vehicleID, req.InvoiceID, req.ServiceJobID, req.Mileage, req.Note, userID).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create record: %w", err)
	}
	return s.getRecord(ctx, branchID, id)
}

// UpdateRecord edits a record's note and mileage (invoice/job links stay).
func (s *Service) UpdateRecord(ctx context.Context, branchID, id int64, req *dto.CreateRecordRequest) (*dto.RecordResponse, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE vehicle_records SET note = NULLIF($1, ''), mileage = $2 WHERE id = $3 AND branch_id = $4`,
		req.Note, req.Mileage, id, branchID)
	if err != nil {
		return nil, fmt.Errorf("update record: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return s.getRecord(ctx, branchID, id)
}

// DeleteRecord removes a record and returns the photo URLs it held, so the
// caller can best-effort clean them out of storage (cascade handles the DB rows).
func (s *Service) DeleteRecord(ctx context.Context, branchID, id int64) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.url FROM vehicle_record_photos p
		JOIN vehicle_records r ON r.id = p.record_id
		WHERE p.record_id = $1 AND r.branch_id = $2`, id, branchID)
	if err != nil {
		return nil, fmt.Errorf("list photo urls: %w", err)
	}
	var urls []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan url: %w", err)
		}
		urls = append(urls, u)
	}
	rows.Close()

	result, err := s.pool.Exec(ctx, `DELETE FROM vehicle_records WHERE id = $1 AND branch_id = $2`, id, branchID)
	if err != nil {
		return nil, fmt.Errorf("delete record: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return urls, nil
}

func (s *Service) AddPhoto(ctx context.Context, branchID, recordID int64, url string) (*dto.PhotoResponse, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vehicle_records WHERE id = $1 AND branch_id = $2)`,
		recordID, branchID).Scan(&exists); err != nil || !exists {
		return nil, domain.ErrNotFound
	}
	var p dto.PhotoResponse
	err := s.pool.QueryRow(ctx,
		`INSERT INTO vehicle_record_photos (record_id, url) VALUES ($1, $2) RETURNING id, url`,
		recordID, url).Scan(&p.ID, &p.URL)
	if err != nil {
		return nil, fmt.Errorf("add photo: %w", err)
	}
	return &p, nil
}

// DeletePhoto removes the DB row and returns the URL so the caller can clean
// up storage.
func (s *Service) DeletePhoto(ctx context.Context, branchID, photoID int64) (string, error) {
	var url string
	err := s.pool.QueryRow(ctx, `
		SELECT p.url FROM vehicle_record_photos p
		JOIN vehicle_records r ON r.id = p.record_id
		WHERE p.id = $1 AND r.branch_id = $2`, photoID, branchID).Scan(&url)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", domain.ErrNotFound
		}
		return "", fmt.Errorf("find photo: %w", err)
	}
	if _, err := s.pool.Exec(ctx, `DELETE FROM vehicle_record_photos WHERE id = $1`, photoID); err != nil {
		return "", fmt.Errorf("delete photo: %w", err)
	}
	return url, nil
}
