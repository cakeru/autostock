package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/vehicle/dto"
)

// Shop-wide defaults when a branch hasn't customized its settings yet.
// Chosen for a Cambodia-typical mixed passenger/pickup fleet: oil every
// 5,000km or 3 months, tires by ~40,000km tread life or 4 years age
// (whichever the car hits first), and a conservative 30km/day fallback for
// estimating "today's odometer" when a vehicle only has one mileage reading
// on file (so a new customer still gets a usable, if soft, estimate).
const (
	defaultOilIntervalKm   = 5000
	defaultOilIntervalDays = 90
	// Fallback tire life (km) for a tire with no rated life on file. Tires are
	// judged purely by wear-to-km now, so there's no default time interval.
	defaultTireLifeKm       = 40000
	defaultFallbackKmPerDay = 30.0
	defaultDueSoonDays      = 14
)

type intervalSettings struct {
	oilKm, oilDays, tireLifeKm, tireDays, dueSoonDays int
	fallbackKmPerDay                                  float64
}

func (s *Service) loadIntervalSettings(ctx context.Context, branchID int64) (intervalSettings, error) {
	iv := intervalSettings{
		oilKm: defaultOilIntervalKm, oilDays: defaultOilIntervalDays,
		tireLifeKm: defaultTireLifeKm,
		fallbackKmPerDay: defaultFallbackKmPerDay, dueSoonDays: defaultDueSoonDays,
	}
	rows, err := s.pool.Query(ctx, `
		SELECT key, value FROM settings WHERE branch_id = $1 AND key IN
		('service_oil_interval_km', 'service_oil_interval_days', 'service_tire_life_km',
		 'service_tire_interval_days', 'service_fallback_km_per_day', 'service_due_soon_days')`, branchID)
	if err != nil {
		return iv, fmt.Errorf("load interval settings: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return iv, fmt.Errorf("scan interval setting: %w", err)
		}
		switch key {
		case "service_oil_interval_km":
			if v, err := strconv.Atoi(value); err == nil {
				iv.oilKm = v
			}
		case "service_oil_interval_days":
			if v, err := strconv.Atoi(value); err == nil {
				iv.oilDays = v
			}
		case "service_tire_life_km":
			if v, err := strconv.Atoi(value); err == nil {
				iv.tireLifeKm = v
			}
		case "service_tire_interval_days":
			if v, err := strconv.Atoi(value); err == nil {
				iv.tireDays = v
			}
		case "service_fallback_km_per_day":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				iv.fallbackKmPerDay = v
			}
		case "service_due_soon_days":
			if v, err := strconv.Atoi(value); err == nil {
				iv.dueSoonDays = v
			}
		}
	}
	return iv, nil
}

func (s *Service) GetIntervalSettings(ctx context.Context, branchID int64) (*dto.IntervalSettingsResponse, error) {
	iv, err := s.loadIntervalSettings(ctx, branchID)
	if err != nil {
		return nil, err
	}
	rules, err := s.loadPartRules(ctx, branchID)
	if err != nil {
		return nil, err
	}
	return &dto.IntervalSettingsResponse{
		OilIntervalKm: iv.oilKm, OilIntervalDays: iv.oilDays,
		TireLifeKm: iv.tireLifeKm, TireIntervalDays: iv.tireDays,
		FallbackKmPerDay: iv.fallbackKmPerDay, DueSoonDays: iv.dueSoonDays,
		PartRules: rules,
	}, nil
}

func (s *Service) UpdateIntervalSettings(ctx context.Context, branchID int64, req *dto.UpdateIntervalSettingsRequest) error {
	set := func(key string, v any) error {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO settings (branch_id, key, value) VALUES ($1, $2, $3)
			ON CONFLICT (branch_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
			branchID, key, fmt.Sprintf("%v", v))
		return err
	}
	if req.OilIntervalKm != nil {
		if err := set("service_oil_interval_km", *req.OilIntervalKm); err != nil {
			return err
		}
	}
	if req.OilIntervalDays != nil {
		if err := set("service_oil_interval_days", *req.OilIntervalDays); err != nil {
			return err
		}
	}
	if req.TireLifeKm != nil {
		if err := set("service_tire_life_km", *req.TireLifeKm); err != nil {
			return err
		}
	}
	if req.TireIntervalDays != nil {
		if err := set("service_tire_interval_days", *req.TireIntervalDays); err != nil {
			return err
		}
	}
	if req.FallbackKmPerDay != nil {
		if err := set("service_fallback_km_per_day", *req.FallbackKmPerDay); err != nil {
			return err
		}
	}
	if req.DueSoonDays != nil {
		if err := set("service_due_soon_days", *req.DueSoonDays); err != nil {
			return err
		}
	}
	if req.PartRules != nil {
		payload, err := json.Marshal(req.PartRules)
		if err != nil {
			return fmt.Errorf("marshal part rules: %w", err)
		}
		if err := set("service_part_reminder_rules", string(payload)); err != nil {
			return err
		}
	}
	return nil
}

// defaultPartRules seed a sensible starting set so the Settings UI isn't blank
// on first open. Admin edits/removes freely; once saved, the stored set wins.
func defaultPartRules() []dto.PartRule {
	km := func(v int) *int { return &v }
	days := func(v int) *int { return &v }
	return []dto.PartRule{
		{PartKey: "air_filter", Label: "Air filter", Km: km(15000)},
		{PartKey: "brakes_front", Label: "Brakes (front)", Km: km(40000)},
		{PartKey: "brakes_rear", Label: "Brakes (rear)", Km: km(40000)},
		{PartKey: "battery", Label: "Battery", Days: days(1095)},
		{PartKey: "wipers", Label: "Wipers", Days: days(365)},
		{PartKey: "coolant", Label: "Coolant", Km: km(40000), Days: days(730)},
	}
}

func (s *Service) loadPartRules(ctx context.Context, branchID int64) ([]dto.PartRule, error) {
	var raw string
	err := s.pool.QueryRow(ctx,
		`SELECT value FROM settings WHERE branch_id = $1 AND key = 'service_part_reminder_rules'`, branchID).Scan(&raw)
	if err != nil {
		if err == pgx.ErrNoRows {
			return defaultPartRules(), nil
		}
		return nil, fmt.Errorf("load part rules: %w", err)
	}
	var rules []dto.PartRule
	if err := json.Unmarshal([]byte(raw), &rules); err != nil {
		return defaultPartRules(), nil
	}
	if rules == nil {
		rules = []dto.PartRule{}
	}
	return rules, nil
}

func (s *Service) UpdateVehicleIntervals(ctx context.Context, branchID, vehicleID int64, req *dto.UpdateVehicleIntervalsRequest) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE vehicles SET oil_interval_km = $1, oil_interval_days = $2,
		    tire_interval_km = $3, tire_interval_days = $4, updated_at = NOW()
		WHERE id = $5 AND branch_id = $6`,
		req.OilIntervalKm, req.OilIntervalDays, req.TireIntervalKm, req.TireIntervalDays, vehicleID, branchID)
	if err != nil {
		return fmt.Errorf("update vehicle intervals: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type mileagePoint struct {
	mileage int
	at      time.Time
}

// mileageHistoryBounds returns the earliest and latest known (mileage, date)
// readings for a vehicle across every source that captures one — invoices,
// service jobs, and vehicle_records — so the velocity estimate reflects the
// car's whole visible history, not just its service events.
func (s *Service) mileageHistoryBounds(ctx context.Context, vehicleID int64) (first, last *mileagePoint, err error) {
	rows, err := s.pool.Query(ctx, `
		WITH points AS (
			SELECT mileage, created_at AS at FROM vehicle_records WHERE vehicle_id = $1 AND mileage IS NOT NULL
			UNION ALL
			SELECT mileage, COALESCE(issued_at, created_at) AS at FROM invoices WHERE vehicle_id = $1 AND mileage IS NOT NULL AND status <> 'voided'
			UNION ALL
			SELECT mileage, COALESCE(completed_at, created_at) AS at FROM service_jobs WHERE vehicle_id = $1 AND mileage IS NOT NULL
			UNION ALL
			SELECT mileage, occurred_at AS at FROM vehicle_service_events WHERE vehicle_id = $1 AND mileage IS NOT NULL
			UNION ALL
			SELECT mileage, performed_at AS at FROM vehicle_wheel_services WHERE vehicle_id = $1 AND mileage IS NOT NULL
			UNION ALL
			SELECT mileage, replaced_at AS at FROM vehicle_parts WHERE vehicle_id = $1 AND mileage IS NOT NULL
		)
		SELECT mileage, at FROM points ORDER BY at ASC`, vehicleID)
	if err != nil {
		return nil, nil, fmt.Errorf("mileage history: %w", err)
	}
	defer rows.Close()

	var points []mileagePoint
	for rows.Next() {
		var p mileagePoint
		if err := rows.Scan(&p.mileage, &p.at); err != nil {
			return nil, nil, fmt.Errorf("scan mileage point: %w", err)
		}
		points = append(points, p)
	}
	if len(points) == 0 {
		return nil, nil, nil
	}
	return &points[0], &points[len(points)-1], nil
}

// mileageVelocity estimates km/day driven from the vehicle's own earliest-to-
// latest mileage reading, falling back to the shop-wide default when there's
// only one reading (or readings share a timestamp) so a new vehicle still
// gets a usable, if soft, projection instead of none at all.
func mileageVelocity(first, last *mileagePoint, fallbackKmPerDay float64) float64 {
	if first != nil && last != nil && first.at.Before(last.at) && last.mileage > first.mileage {
		spanDays := last.at.Sub(first.at).Hours() / 24
		if spanDays >= 1 {
			return float64(last.mileage-first.mileage) / spanDays
		}
	}
	return fallbackKmPerDay
}

// estimateTodayMileage projects the vehicle's current odometer forward from
// its last known reading using mileageVelocity.
func estimateTodayMileage(first, last *mileagePoint, fallbackKmPerDay float64, now time.Time) *int {
	if last == nil {
		return nil
	}
	daysSinceLast := now.Sub(last.at).Hours() / 24
	if daysSinceLast <= 0 {
		v := last.mileage
		return &v
	}
	kmPerDay := mileageVelocity(first, last, fallbackKmPerDay)
	estimated := last.mileage + int(kmPerDay*daysSinceLast)
	return &estimated
}

// reminderInput describes one thing to remind about: where it was last done
// (odometer + date) and its limits — a km life and/or a day interval. Either
// limit may be nil; whichever is reached first (projected onto today via the
// vehicle's mileage velocity) wins.
type reminderInput struct {
	eventType   string // oil | tire | part
	key         string // oil | tire | <part_key>
	label       string
	mileage     *int
	at          *time.Time
	kmLimit     *int
	dayLimit    *int
	monthsLimit *int // oil only: calendar-month interval from the sold product
}

func computeDue(in reminderInput, first, last *mileagePoint, fallbackKmPerDay float64, dueSoonDays int, now time.Time) dto.DueStatus {
	status := dto.DueStatus{
		EventType: in.eventType, Key: in.key, Label: in.label,
		LastMileage: in.mileage, LastServiceAt: in.at, Status: "unknown",
	}

	var dateDue, mileageDue *time.Time
	if in.at != nil && in.dayLimit != nil {
		d := in.at.AddDate(0, 0, *in.dayLimit)
		dateDue = &d
	}
	if in.at != nil && in.monthsLimit != nil {
		d := in.at.AddDate(0, *in.monthsLimit, 0)
		if dateDue == nil || d.Before(*dateDue) {
			dateDue = &d
		}
	}
	if in.mileage != nil && in.kmLimit != nil {
		est := estimateTodayMileage(first, last, fallbackKmPerDay, now)
		status.EstimatedMileageToday = est
		dueMileage := *in.mileage + *in.kmLimit
		status.DueMileage = &dueMileage
		if est != nil {
			kmPerDay := mileageVelocity(first, last, fallbackKmPerDay)
			if kmPerDay <= 0 {
				kmPerDay = fallbackKmPerDay
			}
			daysOut := float64(dueMileage-*est) / kmPerDay
			d := now.Add(time.Duration(daysOut * 24 * float64(time.Hour)))
			mileageDue = &d
		}
	}

	// Whichever limit comes first wins. Record which basis it was: a date-driven
	// due is a certain calendar date; a mileage-driven one is an estimate.
	var due *time.Time
	basis := ""
	if dateDue != nil {
		due, basis = dateDue, "date"
	}
	if mileageDue != nil && (due == nil || mileageDue.Before(*due)) {
		due, basis = mileageDue, "mileage"
	}
	if due == nil {
		return status // unknown — nothing to project from
	}
	status.DueDate = due
	status.DueBasis = basis

	switch {
	case !due.After(now):
		status.Status = "overdue"
	case due.Before(now.AddDate(0, 0, dueSoonDays)):
		status.Status = "due_soon"
	default:
		status.Status = "ok"
	}
	return status
}

func intPtr(v int) *int { return &v }

func partLabel(r dto.PartRule) string {
	if r.Label != "" {
		return r.Label
	}
	return r.PartKey
}

// lastServiceEvent returns the most recent (mileage, date) for an oil/tire
// event plus the interval recorded at the time (life_km distance, life_months
// time — nil when the event didn't capture one), or nils if there's no event.
func (s *Service) lastServiceEvent(ctx context.Context, branchID, vehicleID int64, eventType string) (mileage *int, at *time.Time, lifeKm, lifeDays, lifeMonths *int, err error) {
	err = s.pool.QueryRow(ctx,
		`SELECT mileage, occurred_at, life_km, life_days, life_months FROM vehicle_service_events
		 WHERE vehicle_id = $1 AND branch_id = $2 AND event_type = $3
		 ORDER BY occurred_at DESC LIMIT 1`, vehicleID, branchID, eventType).Scan(&mileage, &at, &lifeKm, &lifeDays, &lifeMonths)
	if err == pgx.ErrNoRows {
		return nil, nil, nil, nil, nil, nil
	}
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("last %s event: %w", eventType, err)
	}
	return mileage, at, lifeKm, lifeDays, lifeMonths, nil
}

// lastTireEvent additionally returns the effective life (km) recorded for the
// most recent tire install (nil when the install didn't capture one).
func (s *Service) lastTireEvent(ctx context.Context, branchID, vehicleID int64) (mileage *int, at *time.Time, lifeKm, lifeDays, lifeMonths *int, err error) {
	err = s.pool.QueryRow(ctx,
		`SELECT mileage, occurred_at, life_km, life_days, life_months FROM vehicle_service_events
		 WHERE vehicle_id = $1 AND branch_id = $2 AND event_type = 'tire'
		 ORDER BY occurred_at DESC LIMIT 1`, vehicleID, branchID).Scan(&mileage, &at, &lifeKm, &lifeDays, &lifeMonths)
	if err == pgx.ErrNoRows {
		return nil, nil, nil, nil, nil, nil
	}
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("last tire event: %w", err)
	}
	return mileage, at, lifeKm, lifeDays, lifeMonths, nil
}

// lastPartReplacement anchors a part reminder on the newest logged replacement
// of that part for the vehicle.
func (s *Service) lastPartReplacement(ctx context.Context, vehicleID int64, partKey string) (*int, *time.Time, error) {
	var mileage *int
	var at *time.Time
	err := s.pool.QueryRow(ctx,
		`SELECT mileage, replaced_at FROM vehicle_parts
		 WHERE vehicle_id = $1 AND part_key = $2
		 ORDER BY replaced_at DESC LIMIT 1`, vehicleID, partKey).Scan(&mileage, &at)
	if err != nil && err != pgx.ErrNoRows {
		return nil, nil, fmt.Errorf("last %s replacement: %w", partKey, err)
	}
	return mileage, at, nil
}

// vehicleIntervalOverrides are the per-vehicle interval columns; any nil field
// falls back to the branch default.
type vehicleIntervalOverrides struct {
	oilKm, oilDays, tireKm, tireDays *int
}

// intPtrIf returns &v when v > 0, else nil — a 0 day/km interval means "no such
// limit" (tires default to km-only until a day interval is configured).
func intPtrIf(v int) *int {
	if v <= 0 {
		return nil
	}
	return &v
}

// reminderInputs assembles every reminder for a vehicle: oil and tire always
// (may be "unknown" when un-anchored), plus one per configured part rule that
// has an actual replacement to anchor on.
func (s *Service) reminderInputs(ctx context.Context, branchID, vehicleID int64, iv intervalSettings, partRules []dto.PartRule, ovr vehicleIntervalOverrides) ([]reminderInput, error) {
	var inputs []reminderInput

	oilMileage, oilAt, oilLifeKm, oilLifeDays, oilLifeMonths, err := s.lastServiceEvent(ctx, branchID, vehicleID, "oil")
	if err != nil {
		return nil, err
	}
	// The sold oil product's own rating (km / days / months) is most
	// authoritative; missing pieces fall back to the per-vehicle override, then
	// the branch default. A product time rating replaces the branch's default.
	oilKm := iv.oilKm
	if ovr.oilKm != nil {
		oilKm = *ovr.oilKm
	}
	if oilLifeKm != nil {
		oilKm = *oilLifeKm
	}
	oilDays := iv.oilDays
	if ovr.oilDays != nil {
		oilDays = *ovr.oilDays
	}
	var dayLimit *int
	var monthsLimit *int
	switch {
	case oilLifeMonths != nil:
		monthsLimit = oilLifeMonths
		dayLimit = oilLifeDays
	case oilLifeDays != nil:
		dayLimit = oilLifeDays
	default:
		dayLimit = intPtrIf(oilDays)
	}
	inputs = append(inputs, reminderInput{
		eventType: "oil", key: "oil", label: "Oil change",
		mileage: oilMileage, at: oilAt, kmLimit: intPtrIf(oilKm), dayLimit: dayLimit, monthsLimit: monthsLimit,
	})

	tireMileage, tireAt, tireLifeKm, tireLifeDays, tireLifeMonths, err := s.lastTireEvent(ctx, branchID, vehicleID)
	if err != nil {
		return nil, err
	}
	// km life: the actually-installed tire's own rating is most authoritative,
	// then the per-vehicle override, then the branch default.
	life := iv.tireLifeKm
	if ovr.tireKm != nil {
		life = *ovr.tireKm
	}
	if tireLifeKm != nil {
		life = *tireLifeKm
	}
	// Time: an installed tire's day/month rating, else the branch default
	// (off by default → km-only). A month rating replaces the day default.
	tireDays := iv.tireDays
	if ovr.tireDays != nil {
		tireDays = *ovr.tireDays
	}
	var tireDayLimit *int
	var tireMonthsLimit *int
	switch {
	case tireLifeMonths != nil:
		tireMonthsLimit = tireLifeMonths
		tireDayLimit = tireLifeDays
	case tireLifeDays != nil:
		tireDayLimit = tireLifeDays
	default:
		tireDayLimit = intPtrIf(tireDays)
	}
	inputs = append(inputs, reminderInput{
		eventType: "tire", key: "tire", label: "Tires",
		mileage: tireMileage, at: tireAt, kmLimit: intPtrIf(life), dayLimit: tireDayLimit, monthsLimit: tireMonthsLimit,
	})

	for _, rule := range partRules {
		if rule.Km == nil && rule.Days == nil {
			continue
		}
		pm, pat, err := s.lastPartReplacement(ctx, vehicleID, rule.PartKey)
		if err != nil {
			return nil, err
		}
		if pat == nil {
			continue // never replaced through us — nothing to anchor on
		}
		inputs = append(inputs, reminderInput{
			eventType: "part", key: rule.PartKey, label: partLabel(rule),
			mileage: pm, at: pat, kmLimit: rule.Km, dayLimit: rule.Days,
		})
	}
	return inputs, nil
}

// GetDueForVehicle returns both event types' due status for one vehicle
// (used on the vehicle profile page) — including "unknown" entries, so the
// page can show "no oil history yet" rather than silently omitting a row.
func (s *Service) GetDueForVehicle(ctx context.Context, branchID, vehicleID int64) ([]dto.DueStatus, error) {
	iv, err := s.loadIntervalSettings(ctx, branchID)
	if err != nil {
		return nil, err
	}
	partRules, err := s.loadPartRules(ctx, branchID)
	if err != nil {
		return nil, err
	}

	var ovr vehicleIntervalOverrides
	err = s.pool.QueryRow(ctx,
		`SELECT oil_interval_km, oil_interval_days, tire_interval_km, tire_interval_days
		 FROM vehicles WHERE id = $1 AND branch_id = $2`,
		vehicleID, branchID).Scan(&ovr.oilKm, &ovr.oilDays, &ovr.tireKm, &ovr.tireDays)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("load vehicle intervals: %w", err)
	}

	first, last, err := s.mileageHistoryBounds(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	inputs, err := s.reminderInputs(ctx, branchID, vehicleID, iv, partRules, ovr)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	out := make([]dto.DueStatus, 0, len(inputs))
	for _, in := range inputs {
		out = append(out, computeDue(in, first, last, iv.fallbackKmPerDay, iv.dueSoonDays, now))
	}
	return out, nil
}

// ListDueForService returns every vehicle in the branch that's overdue or
// due soon for oil or tires — the shop-wide "who needs a call" list. Vehicles
// with no detected service history are skipped entirely rather than guessed
// at, since there's nothing to project from.
func (s *Service) ListDueForService(ctx context.Context, branchID int64, horizonDays int) ([]dto.DueForServiceItem, error) {
	iv, err := s.loadIntervalSettings(ctx, branchID)
	if err != nil {
		return nil, err
	}
	partRules, err := s.loadPartRules(ctx, branchID)
	if err != nil {
		return nil, err
	}

	// Any vehicle with something to project from — a service event, a logged
	// part replacement, or a wheel reading.
	rows, err := s.pool.Query(ctx, `
		SELECT vehicle_id FROM vehicle_service_events WHERE branch_id = $1
		UNION
		SELECT vehicle_id FROM vehicle_parts WHERE branch_id = $1 AND part_key IS NOT NULL
		UNION
		SELECT vehicle_id FROM vehicle_wheel_services WHERE branch_id = $1`, branchID)
	if err != nil {
		return nil, fmt.Errorf("list vehicles with history: %w", err)
	}
	var vehicleIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan vehicle id: %w", err)
		}
		vehicleIDs = append(vehicleIDs, id)
	}
	rows.Close()

	now := time.Now()
	out := []dto.DueForServiceItem{}
	for _, vehicleID := range vehicleIDs {
		var plate, make_, model, custName, custPhone string
		var custID int64
		var ovr vehicleIntervalOverrides
		err := s.pool.QueryRow(ctx, `
			SELECT vh.plate_number, COALESCE(vh.make,''), COALESCE(vh.model,''), vh.customer_id,
			       COALESCE(c.name,''), COALESCE(c.phone,''),
			       vh.oil_interval_km, vh.oil_interval_days, vh.tire_interval_km, vh.tire_interval_days
			FROM vehicles vh JOIN customers c ON c.id = vh.customer_id
			WHERE vh.id = $1 AND vh.branch_id = $2`, vehicleID, branchID).
			Scan(&plate, &make_, &model, &custID, &custName, &custPhone,
				&ovr.oilKm, &ovr.oilDays, &ovr.tireKm, &ovr.tireDays)
		if err != nil {
			continue
		}

		first, last, err := s.mileageHistoryBounds(ctx, vehicleID)
		if err != nil {
			continue
		}

		inputs, err := s.reminderInputs(ctx, branchID, vehicleID, iv, partRules, ovr)
		if err != nil {
			continue
		}
		for _, in := range inputs {
			ds := computeDue(in, first, last, iv.fallbackKmPerDay, iv.dueSoonDays, now)
			// The list (call sheet) is overdue + due-soon only. The calendar asks
			// for a forward horizon too, so on-track items with a due date inside
			// the window are included so the month grid isn't empty.
			include := ds.Status == "overdue" || ds.Status == "due_soon"
			if !include && horizonDays > 0 && ds.Status == "ok" && ds.DueDate != nil &&
				!ds.DueDate.After(now.AddDate(0, 0, horizonDays)) {
				include = true
			}
			if include {
				out = append(out, dto.DueForServiceItem{
					VehicleID: vehicleID, PlateNumber: plate, Make: make_, Model: model,
					CustomerID: custID, CustomerName: custName, CustomerPhone: custPhone,
					DueStatus: ds,
				})
			}
		}
	}

	return out, nil
}

// ---------------------------------------------------------------------------
// Manual service event management (corrections / backfilling pre-feature history)
// ---------------------------------------------------------------------------

func (s *Service) ListServiceEvents(ctx context.Context, branchID, vehicleID int64) ([]dto.ServiceEventResponse, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT e.id, e.event_type, e.mileage, e.occurred_at, e.invoice_id, COALESCE(i.invoice_number,''),
		       COALESCE(e.product_name,''), COALESCE(u.full_name,'')
		FROM vehicle_service_events e
		LEFT JOIN invoices i ON i.id = e.invoice_id
		LEFT JOIN users u ON u.id = e.created_by
		WHERE e.vehicle_id = $1 AND e.branch_id = $2
		ORDER BY e.occurred_at DESC`, vehicleID, branchID)
	if err != nil {
		return nil, fmt.Errorf("list service events: %w", err)
	}
	defer rows.Close()

	events := []dto.ServiceEventResponse{}
	for rows.Next() {
		var e dto.ServiceEventResponse
		if err := rows.Scan(&e.ID, &e.EventType, &e.Mileage, &e.OccurredAt, &e.InvoiceID, &e.InvoiceNumber,
			&e.ProductName, &e.CreatedByName); err != nil {
			return nil, fmt.Errorf("scan service event: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}

func (s *Service) CreateServiceEvent(ctx context.Context, branchID, vehicleID, userID int64, req *dto.CreateServiceEventRequest) (*dto.ServiceEventResponse, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND branch_id = $2)`,
		vehicleID, branchID).Scan(&exists); err != nil || !exists {
		return nil, domain.ErrNotFound
	}

	occurredAt := time.Now()
	if req.OccurredAt != "" {
		t, err := time.Parse("2006-01-02", req.OccurredAt)
		if err != nil {
			return nil, &domain.AppError{Code: "INVALID_DATE", Message: "occurred_at must be YYYY-MM-DD", Status: 400}
		}
		occurredAt = t
	}

	lifeKm := req.LifeKm
	if req.EventType != "tire" {
		lifeKm = nil // life_km only meaningful for tire installs
	}

	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO vehicle_service_events (branch_id, vehicle_id, event_type, mileage, occurred_at, product_name, life_km, created_by)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8)
		RETURNING id`,
		branchID, vehicleID, req.EventType, req.Mileage, occurredAt, req.ProductName, lifeKm, userID).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create service event: %w", err)
	}

	var e dto.ServiceEventResponse
	err = s.pool.QueryRow(ctx, `
		SELECT e.id, e.event_type, e.mileage, e.occurred_at, e.invoice_id, COALESCE(i.invoice_number,''),
		       COALESCE(e.product_name,''), COALESCE(u.full_name,'')
		FROM vehicle_service_events e
		LEFT JOIN invoices i ON i.id = e.invoice_id
		LEFT JOIN users u ON u.id = e.created_by
		WHERE e.id = $1`, id).
		Scan(&e.ID, &e.EventType, &e.Mileage, &e.OccurredAt, &e.InvoiceID, &e.InvoiceNumber, &e.ProductName, &e.CreatedByName)
	if err != nil {
		return nil, fmt.Errorf("get created event: %w", err)
	}
	return &e, nil
}

func (s *Service) DeleteServiceEvent(ctx context.Context, branchID, eventID int64) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM vehicle_service_events WHERE id = $1 AND branch_id = $2`, eventID, branchID)
	if err != nil {
		return fmt.Errorf("delete service event: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
