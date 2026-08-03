package service

import (
	"context"
	"sort"
	"time"

	"github.com/cakeru/autostock/internal/vehicle/dto"
)

// clusterWindow is how close two dated items must be to count as the same visit.
// Work booked on a Friday and finished the Monday after still reads as one visit;
// separate visits are normally weeks apart.
const clusterWindow = 2 * 24 * time.Hour

// GetTimeline assembles the vehicle's whole service history as a list of visits,
// newest first, for the internal (authenticated) vehicle page — everything that
// happened each day: what was done, what it cost, and the photo evidence.
func (s *Service) GetTimeline(ctx context.Context, branchID, vehicleID int64) ([]dto.VisitResponse, error) {
	return s.buildTimeline(ctx, branchID, vehicleID, false)
}

// GetPublicTimeline is the customer-safe cut of the same history: no prices or
// sale/job links, no internal notes/photos — only the work done (tires,
// alignment, oil, parts) plus photos the shop explicitly shared. Visits that
// would be empty after that filtering are dropped.
func (s *Service) GetPublicTimeline(ctx context.Context, branchID, vehicleID int64) ([]dto.VisitResponse, error) {
	return s.buildTimeline(ctx, branchID, vehicleID, true)
}

// buildTimeline gathers every dated source and clusters by date into visits.
// publicOnly strips everything a customer shouldn't see.
func (s *Service) buildTimeline(ctx context.Context, branchID, vehicleID int64, publicOnly bool) ([]dto.VisitResponse, error) {
	wheels, err := s.ListWheelServices(ctx, branchID, vehicleID)
	if err != nil {
		return nil, err
	}
	events, err := s.ListServiceEvents(ctx, branchID, vehicleID)
	if err != nil {
		return nil, err
	}
	parts, err := s.ListParts(ctx, branchID, vehicleID)
	if err != nil {
		return nil, err
	}
	photos, err := s.ListGalleryPhotos(ctx, branchID, vehicleID)
	if err != nil {
		return nil, err
	}
	// Transactions (prices), free-form records (internal notes), and batch-scan
	// installs (internal batch/DOT traceability) are all internal only — the
	// customer report never sees them.
	var history []dto.HistoryItem
	var records []dto.RecordResponse
	var installs []installRow
	if !publicOnly {
		if history, err = s.GetHistory(ctx, branchID, vehicleID); err != nil {
			return nil, err
		}
		if records, err = s.ListRecords(ctx, branchID, vehicleID); err != nil {
			return nil, err
		}
		if installs, err = s.listInstallRows(ctx, branchID, vehicleID); err != nil {
			return nil, err
		}
	}

	const (
		kWheel = iota
		kEvent
		kPart
		kTxn
		kRecord
		kPhoto
		kInstall
	)
	type item struct {
		at   time.Time
		kind int
		idx  int
	}
	var items []item
	for i, w := range wheels {
		items = append(items, item{w.PerformedAt, kWheel, i})
	}
	for i, e := range events {
		items = append(items, item{e.OccurredAt, kEvent, i})
	}
	for i, p := range parts {
		items = append(items, item{p.ReplacedAt, kPart, i})
	}
	for i, h := range history {
		if h.Date != nil {
			items = append(items, item{*h.Date, kTxn, i})
		}
	}
	for i, r := range records {
		items = append(items, item{r.CreatedAt, kRecord, i})
	}
	for i, ph := range photos {
		if publicOnly && !ph.CustomerVisible {
			continue // customer sees only the photos the shop shared
		}
		items = append(items, item{ph.TakenAt, kPhoto, i})
	}
	for i, in := range installs {
		items = append(items, item{in.installedAt, kInstall, i})
	}
	if len(items) == 0 {
		return []dto.VisitResponse{}, nil
	}

	sort.Slice(items, func(a, b int) bool { return items[a].at.After(items[b].at) })

	// Chain items into visits: an item joins the current visit when it's within
	// the window of the previous (chronologically adjacent) item.
	visits := []dto.VisitResponse{}
	var cur []item
	flush := func() {
		if len(cur) == 0 {
			return
		}
		v := dto.VisitResponse{Date: cur[0].at} // newest item in the cluster
		var mileage *int
		// Lower priority number wins; the wheel service's odometer is the most
		// authoritative, then an event, then a part/record, then an invoice.
		milePrio := 99
		consider := func(m *int, prio int) {
			if m != nil && prio < milePrio {
				mileage, milePrio = m, prio
			}
		}
		for _, it := range cur {
			switch it.kind {
			case kWheel:
				w := wheels[it.idx]
				if v.WheelService == nil {
					ws := w
					v.WheelService = &ws
				}
				consider(w.Mileage, 0)
				if w.Notes != "" {
					v.Notes = append(v.Notes, w.Notes)
				}
				// Alignment printouts stay internal — only shared gallery photos
				// reach the customer report.
				if !publicOnly {
					for _, ph := range w.Photos {
						v.Photos = append(v.Photos, dto.VisitPhoto{URL: ph.URL, Source: "wheel"})
					}
				}
			case kEvent:
				e := events[it.idx]
				consider(e.Mileage, 1)
				if e.EventType == "oil" {
					v.OilChange = true
					if e.ProductName != "" {
						v.OilNote = e.ProductName
					}
					if !publicOnly {
						id := e.ID
						v.OilEventID = &id
					}
				} else if e.EventType == "tire" {
					v.TireChange = true
					if e.ProductName != "" {
						v.TireNote = e.ProductName
					}
					if !publicOnly {
						id := e.ID
						v.TireEventID = &id
					}
				} else if e.EventType == "service" {
					v.Services = append(v.Services, dto.VisitService{ID: e.ID, Name: e.ProductName})
				}
			case kPart:
				v.Parts = append(v.Parts, parts[it.idx])
				consider(parts[it.idx].Mileage, 2)
			case kTxn:
				h := history[it.idx]
				v.Transactions = append(v.Transactions, dto.VisitTxn{
					Type: h.Type, ID: h.ID, Ref: h.Ref, Amount: h.Amount, Status: h.Status,
				})
				consider(h.Mileage, 3)
			case kRecord:
				r := records[it.idx]
				consider(r.Mileage, 2)
				if r.Note != "" {
					v.Notes = append(v.Notes, r.Note)
				}
				for _, ph := range r.Photos {
					v.Photos = append(v.Photos, dto.VisitPhoto{URL: ph.URL, Source: "record"})
				}
			case kPhoto:
				ph := photos[it.idx]
				v.Photos = append(v.Photos, dto.VisitPhoto{
					ID: ph.ID, URL: ph.URL, Caption: ph.Caption, Phase: ph.Phase,
					Source: "gallery", CustomerVisible: ph.CustomerVisible,
				})
			case kInstall:
				in := installs[it.idx]
				v.Installs = append(v.Installs, dto.VisitInstall{
					BatchNo: in.batchNo, ProductName: in.productName, TireSize: in.tireSize,
					DOTCode: in.dotCode, Position: in.position, MechanicName: in.mechanicName,
				})
			}
		}
		v.Mileage = mileage
		// In the customer view a cluster can end up empty once prices, internal
		// notes and unshared photos are stripped (e.g. a sale-only day) — skip it.
		if publicOnly && v.WheelService == nil && !v.OilChange && !v.TireChange && len(v.Services) == 0 && len(v.Parts) == 0 && len(v.Photos) == 0 {
			cur = nil
			return
		}
		visits = append(visits, v)
		cur = nil
	}

	// Compare each item to the cluster's newest item (its anchor), not the
	// previous one, so a run of near-daily items can't chain-drift into one
	// oversized visit — each visit spans at most the window.
	cur = append(cur, items[0])
	for i := 1; i < len(items); i++ {
		anchor := cur[0]
		if anchor.at.Sub(items[i].at) <= clusterWindow {
			cur = append(cur, items[i])
		} else {
			flush()
			cur = append(cur, items[i])
		}
	}
	flush()

	return visits, nil
}

// installRow is a scanned batch fitted to this car (internal traceability),
// queried straight from batch_installs to keep the timeline self-contained.
type installRow struct {
	installedAt  time.Time
	batchNo      string
	productName  string
	tireSize     string
	dotCode      string
	position     string
	mechanicName string
}

func (s *Service) listInstallRows(ctx context.Context, branchID, vehicleID int64) ([]installRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT bi.installed_at,
		       'B-' || to_char(b.received_at, 'YYYY') || '-' || lpad(b.id::text, 4, '0'),
		       p.name, COALESCE(p.tire_size,''), COALESCE(b.dot_code,''),
		       COALESCE(bi.position,''), COALESCE(e.name,'')
		FROM batch_installs bi
		JOIN batches b ON b.id = bi.batch_id
		JOIN products p ON p.id = bi.product_id
		LEFT JOIN employees e ON e.id = bi.mechanic_employee_id
		WHERE bi.vehicle_id = $1 AND bi.branch_id = $2
		ORDER BY bi.installed_at DESC, bi.id DESC`, vehicleID, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []installRow
	for rows.Next() {
		var r installRow
		if err := rows.Scan(&r.installedAt, &r.batchNo, &r.productName, &r.tireSize, &r.dotCode, &r.position, &r.mechanicName); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}
