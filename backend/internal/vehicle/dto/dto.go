package dto

import "time"

type VehicleProfileResponse struct {
	ID            int64      `json:"id"`
	PlateNumber   string     `json:"plate_number"`
	Make          string     `json:"make,omitempty"`
	Model         string     `json:"model,omitempty"`
	Year          int        `json:"year,omitempty"`
	VIN           string     `json:"vin,omitempty"`
	Color         string     `json:"color,omitempty"`
	BodyType      string     `json:"body_type,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	CustomerID    int64      `json:"customer_id"`
	CustomerName  string     `json:"customer_name"`
	CustomerPhone string     `json:"customer_phone,omitempty"`
	RecordCount   int        `json:"record_count"`
	LastMileage   *int       `json:"last_mileage,omitempty"`
	LastServiceAt *time.Time `json:"last_service_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`

	OilIntervalKm    *int `json:"oil_interval_km,omitempty"`
	OilIntervalDays  *int `json:"oil_interval_days,omitempty"`
	TireIntervalKm   *int `json:"tire_interval_km,omitempty"`
	TireIntervalDays *int `json:"tire_interval_days,omitempty"`

	ShareToken string `json:"share_token,omitempty"`

	Due []DueStatus `json:"due"`
}

// HistoryItem mirrors the customer activity timeline shape (same UI can
// render either), but scoped to one vehicle and mileage-aware.
type HistoryItem struct {
	Type    string     `json:"type"` // "job" | "invoice"
	ID      int64      `json:"id"`
	Ref     string     `json:"ref"`
	Date    *time.Time `json:"date,omitempty"`
	Title   string     `json:"title"`
	Status  string     `json:"status"`
	Amount  float64    `json:"amount"`
	Mileage *int       `json:"mileage,omitempty"`
}

type CreateRecordRequest struct {
	Note         string `json:"note"`
	Mileage      *int   `json:"mileage,omitempty"`
	InvoiceID    *int64 `json:"invoice_id,omitempty"`
	ServiceJobID *int64 `json:"service_job_id,omitempty"`
}

type PhotoResponse struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

type RecordResponse struct {
	ID            int64           `json:"id"`
	Note          string          `json:"note,omitempty"`
	Mileage       *int            `json:"mileage,omitempty"`
	InvoiceID     *int64          `json:"invoice_id,omitempty"`
	InvoiceNumber string          `json:"invoice_number,omitempty"`
	ServiceJobID  *int64          `json:"service_job_id,omitempty"`
	JobNumber     string          `json:"job_number,omitempty"`
	CreatedByName string          `json:"created_by_name,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	Photos        []PhotoResponse `json:"photos"`
}

// ---------------------------------------------------------------------------
// Service reminders (oil / tire due-for-service)
// ---------------------------------------------------------------------------

type ServiceEventResponse struct {
	ID            int64     `json:"id"`
	EventType     string    `json:"event_type"` // oil | tire
	Mileage       *int      `json:"mileage,omitempty"`
	OccurredAt    time.Time `json:"occurred_at"`
	InvoiceID     *int64    `json:"invoice_id,omitempty"`
	InvoiceNumber string    `json:"invoice_number,omitempty"`
	ProductName   string    `json:"product_name,omitempty"`
	CreatedByName string    `json:"created_by_name,omitempty"`
}

type CreateServiceEventRequest struct {
	EventType   string `json:"event_type" binding:"required,oneof=oil tire"`
	Mileage     *int   `json:"mileage,omitempty"`
	OccurredAt  string `json:"occurred_at,omitempty"` // YYYY-MM-DD; defaults to today
	ProductName string `json:"product_name,omitempty"`
	LifeKm      *int   `json:"life_km,omitempty"` // tire only: km life for this install
}

// DueStatus is computed per reminder for one vehicle: where it stands against
// its km/day limits, projected forward from the vehicle's mileage history.
// EventType is the broad kind (oil | tire | part); Key/Label identify the
// specific reminder (Key = "oil" | "tire" | a part_key).
type DueStatus struct {
	EventType             string     `json:"event_type"`
	Key                   string     `json:"key"`
	Label                 string     `json:"label"`
	LastMileage           *int       `json:"last_mileage,omitempty"`
	LastServiceAt         *time.Time `json:"last_service_at,omitempty"`
	EstimatedMileageToday *int       `json:"estimated_mileage_today,omitempty"`
	DueMileage            *int       `json:"due_mileage,omitempty"`
	DueDate               *time.Time `json:"due_date,omitempty"`
	// DueBasis says which limit produced DueDate: "date" = a fixed day interval
	// (a CERTAIN date that won't move) or "mileage" = the drifting km projection.
	// Lets the calendar pin certain dues firmly and mark estimated ones softly.
	DueBasis string `json:"due_basis,omitempty"`
	Status   string `json:"status"` // overdue | due_soon | ok | unknown
}

// DueForServiceItem is one row of the shop-wide "who needs a call" list.
type DueForServiceItem struct {
	VehicleID     int64  `json:"vehicle_id"`
	PlateNumber   string `json:"plate_number"`
	Make          string `json:"make,omitempty"`
	Model         string `json:"model,omitempty"`
	CustomerID    int64  `json:"customer_id"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone,omitempty"`
	DueStatus
}

// ---------------------------------------------------------------------------
// Wheel services — per-corner tire / tread / alignment snapshots
// ---------------------------------------------------------------------------

// CornerData is one wheel position within a snapshot. All fields optional so a
// tire-only or alignment-only job fills just what it touched.
type CornerData struct {
	Position      string   `json:"position"` // FL | FR | RL | RR | SPARE
	TireProductID *int64   `json:"tire_product_id,omitempty"`
	TireBrand     string   `json:"tire_brand,omitempty"`
	TireSize      string   `json:"tire_size,omitempty"`
	TireDOT       string   `json:"tire_dot,omitempty"`
	TreadMM       *float64 `json:"tread_mm,omitempty"`
	TreadBeforeMM *float64 `json:"tread_before_mm,omitempty"`
	Pressure      *float64 `json:"pressure,omitempty"`
	CamberBefore  string   `json:"camber_before,omitempty"`
	CamberAfter   string   `json:"camber_after,omitempty"`
	CasterBefore  string   `json:"caster_before,omitempty"`
	CasterAfter   string   `json:"caster_after,omitempty"`
	ToeBefore     string   `json:"toe_before,omitempty"`
	ToeAfter      string   `json:"toe_after,omitempty"`
	WearNote      string   `json:"wear_note,omitempty"`
}

type WheelServiceResponse struct {
	ID            int64           `json:"id"`
	PerformedAt   time.Time       `json:"performed_at"`
	Mileage       *int            `json:"mileage,omitempty"`
	InvoiceID     *int64          `json:"invoice_id,omitempty"`
	InvoiceNumber string          `json:"invoice_number,omitempty"`
	ServiceJobID  *int64          `json:"service_job_id,omitempty"`
	JobNumber     string          `json:"job_number,omitempty"`
	Notes         string          `json:"notes,omitempty"`
	CreatedByName string          `json:"created_by_name,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	Corners       []CornerData    `json:"corners"`
	Photos        []PhotoResponse `json:"photos"`
}

type CreateWheelServiceRequest struct {
	PerformedAt  string       `json:"performed_at,omitempty"` // YYYY-MM-DD; defaults to today
	Mileage      *int         `json:"mileage,omitempty"`
	InvoiceID    *int64       `json:"invoice_id,omitempty"`
	ServiceJobID *int64       `json:"service_job_id,omitempty"`
	Notes        string       `json:"notes,omitempty"`
	Corners      []CornerData `json:"corners" binding:"required,min=1,dive"`
}

// TireOption is a recent tire purchase offered as a pick when filling a corner,
// so mounting a just-sold tire is one tap instead of retyping brand/size.
type TireOption struct {
	ProductID     int64      `json:"product_id"`
	Name          string     `json:"name"`
	Size          string     `json:"size,omitempty"`
	InvoiceID     *int64     `json:"invoice_id,omitempty"`
	InvoiceNumber string     `json:"invoice_number,omitempty"`
	PurchasedAt   *time.Time `json:"purchased_at,omitempty"`
}

// ---------------------------------------------------------------------------
// General parts-replaced log
// ---------------------------------------------------------------------------

type PartResponse struct {
	ID            int64     `json:"id"`
	PartName      string    `json:"part_name"`
	PartKey       string    `json:"part_key,omitempty"`
	Position      string    `json:"position,omitempty"`
	ReplacedAt    time.Time `json:"replaced_at"`
	Mileage       *int      `json:"mileage,omitempty"`
	ProductID     *int64    `json:"product_id,omitempty"`
	InvoiceID     *int64    `json:"invoice_id,omitempty"`
	InvoiceNumber string    `json:"invoice_number,omitempty"`
	CreatedByName string    `json:"created_by_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreatePartRequest struct {
	PartName   string `json:"part_name" binding:"required"`
	PartKey    string `json:"part_key,omitempty"` // ties to a reminder rule; blank = no reminder
	Position   string `json:"position,omitempty"`
	ReplacedAt string `json:"replaced_at,omitempty"` // YYYY-MM-DD; defaults to today
	Mileage    *int   `json:"mileage,omitempty"`
	ProductID  *int64 `json:"product_id,omitempty"`
	InvoiceID  *int64 `json:"invoice_id,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// DVI part condition (green/yellow/red per part; grey = no row)
// ---------------------------------------------------------------------------

type PartStatusResponse struct {
	PartKey       string    `json:"part_key"`
	Status        string    `json:"status"` // green | yellow | red | grey
	Note          string    `json:"note,omitempty"`
	UpdatedByName string    `json:"updated_by_name,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SetPartStatusRequest struct {
	PartKey string `json:"part_key" binding:"required"`
	Status  string `json:"status" binding:"required,oneof=green yellow red grey"`
	Note    string `json:"note,omitempty"`
}

// ---------------------------------------------------------------------------
// Job photo gallery (captioned, optional before/after — internal only)
// ---------------------------------------------------------------------------

type GalleryPhotoResponse struct {
	ID            int64     `json:"id"`
	URL           string    `json:"url"`
	Caption       string    `json:"caption,omitempty"`
	Phase         string    `json:"phase,omitempty"` // before | after | ""
	CreatedByName   string    `json:"created_by_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	TakenAt         time.Time `json:"taken_at"` // COALESCE(taken_at, created_at) — the day the work was done
	CustomerVisible bool      `json:"customer_visible"`
}

// ---------------------------------------------------------------------------
// Service timeline — one entry per visit, assembled from every dated source
// (wheel services, oil/tire events, parts, invoices/jobs, records, photos)
// clustered by date. Replaces the separate wheel-history and activity lists.
// ---------------------------------------------------------------------------

type VisitPhoto struct {
	ID              int64  `json:"id,omitempty"` // gallery photo id (for the share toggle); 0 for wheel/record
	URL             string `json:"url"`
	Caption         string `json:"caption,omitempty"`
	Phase           string `json:"phase,omitempty"`  // before | after | ""
	Source          string `json:"source,omitempty"` // wheel | record | gallery
	CustomerVisible bool   `json:"customer_visible,omitempty"`
}

type VisitTxn struct {
	Type   string  `json:"type"` // invoice | job
	ID     int64   `json:"id"`
	Ref    string  `json:"ref"`
	Amount float64 `json:"amount"`
	Status string  `json:"status,omitempty"`
}

// VisitInstall is one scanned batch fitted to the car that visit (internal
// only — batch/DOT traceability the customer report doesn't carry).
type VisitInstall struct {
	BatchNo      string `json:"batch_no"`
	ProductName  string `json:"product_name"`
	TireSize     string `json:"tire_size,omitempty"`
	DOTCode      string `json:"dot_code,omitempty"`
	Position     string `json:"position,omitempty"`
	MechanicName string `json:"mechanic_name,omitempty"`
}

type VisitResponse struct {
	Date         time.Time             `json:"date"`
	Mileage      *int                  `json:"mileage,omitempty"`
	WheelService *WheelServiceResponse `json:"wheel_service,omitempty"`
	OilChange    bool                  `json:"oil_change,omitempty"`
	OilNote      string                `json:"oil_note,omitempty"`
	OilEventID   *int64                `json:"oil_event_id,omitempty"`  // internal only — lets the timeline delete a mis-logged event
	TireChange   bool                  `json:"tire_change,omitempty"`   // a tire install logged without per-corner detail
	TireNote     string                `json:"tire_note,omitempty"`
	TireEventID  *int64                `json:"tire_event_id,omitempty"` // internal only
	Installs     []VisitInstall        `json:"installs,omitempty"`
	Parts        []PartResponse        `json:"parts,omitempty"`
	Notes        []string              `json:"notes,omitempty"`
	Transactions []VisitTxn            `json:"transactions,omitempty"`
	Photos       []VisitPhoto          `json:"photos,omitempty"`
}

type UpdateGalleryPhotoRequest struct {
	Caption         *string `json:"caption,omitempty"`
	Phase           *string `json:"phase,omitempty"` // "before" | "after" | "" (clears)
	CustomerVisible *bool   `json:"customer_visible,omitempty"`
}

// ---------------------------------------------------------------------------
// Customer-facing share link + public report
// ---------------------------------------------------------------------------

type ShareLinkResponse struct {
	Token string `json:"token"`
}

// PublicReportResponse is everything the unauthenticated customer report page
// renders. Deliberately excludes internal notes, customer contact details, and
// any row ids — the token holder sees their car's condition, nothing else.
type PublicReportResponse struct {
	ShopName    string `json:"shop_name,omitempty"`
	ShopPhone   string `json:"shop_phone,omitempty"`
	ShopAddress string `json:"shop_address,omitempty"`

	PlateNumber  string `json:"plate_number"`
	Make         string `json:"make,omitempty"`
	Model        string `json:"model,omitempty"`
	Year         int    `json:"year,omitempty"`
	BodyType     string `json:"body_type,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
	DistanceUnit string `json:"distance_unit"` // "km" | "mi" — how the report labels mileages

	GeneratedAt   time.Time              `json:"generated_at"`
	Due           []DueStatus            `json:"due"`
	WheelServices []WheelServiceResponse `json:"wheel_services"`
	PartStatuses  []PartStatusResponse   `json:"part_statuses"`
	Parts         []PartResponse         `json:"parts"`
	Visits        []VisitResponse        `json:"visits"` // customer-safe service history (no prices, only shared photos)
}

// PartRule is one branch-level part reminder: a controlled part_key with an
// optional km interval and/or day interval (either may be nil; a rule with
// neither is inert). Stored as JSON in settings.
type PartRule struct {
	PartKey string `json:"part_key"`
	Label   string `json:"label,omitempty"`
	Km      *int   `json:"km,omitempty"`
	Days    *int   `json:"days,omitempty"`
}

type IntervalSettingsResponse struct {
	OilIntervalKm    int        `json:"oil_interval_km"`
	OilIntervalDays  int        `json:"oil_interval_days"`
	TireLifeKm       int        `json:"tire_life_km"`
	TireIntervalDays int        `json:"tire_interval_days"` // 0 = tires judged by km only
	FallbackKmPerDay float64    `json:"fallback_km_per_day"`
	DueSoonDays      int        `json:"due_soon_days"`
	PartRules        []PartRule `json:"part_rules"`
}

type UpdateIntervalSettingsRequest struct {
	OilIntervalKm    *int       `json:"oil_interval_km,omitempty"`
	OilIntervalDays  *int       `json:"oil_interval_days,omitempty"`
	TireLifeKm       *int       `json:"tire_life_km,omitempty"`
	TireIntervalDays *int       `json:"tire_interval_days,omitempty"`
	FallbackKmPerDay *float64   `json:"fallback_km_per_day,omitempty"`
	DueSoonDays      *int       `json:"due_soon_days,omitempty"`
	PartRules        []PartRule `json:"part_rules,omitempty"`
}

// UpdateVehicleIntervalsRequest lets one vehicle override the branch defaults
// (nil = fall back to the branch setting).
type UpdateVehicleIntervalsRequest struct {
	OilIntervalKm    *int `json:"oil_interval_km"`
	OilIntervalDays  *int `json:"oil_interval_days"`
	TireIntervalKm   *int `json:"tire_interval_km"`
	TireIntervalDays *int `json:"tire_interval_days"`
}
