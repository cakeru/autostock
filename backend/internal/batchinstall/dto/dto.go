package dto

import "time"

// BatchInfo is what a scanned QR resolves to — enough for the mechanic to
// confirm they grabbed the right lot before recording the install.
type BatchInfo struct {
	BatchID           int64     `json:"batch_id"`
	BatchNo           string    `json:"batch_no"`
	ProductID         int64     `json:"product_id"`
	ProductName       string    `json:"product_name"`
	TireSize          string    `json:"tire_size,omitempty"`
	DOTCode           string    `json:"dot_code,omitempty"`
	Supplier          string    `json:"supplier,omitempty"`
	QuantityRemaining float64   `json:"quantity_remaining"`
	ReceivedAt        time.Time `json:"received_at"`
}

type RecordInstallRequest struct {
	BatchID            int64  `json:"batch_id" binding:"required"`
	ServiceJobID       int64  `json:"service_job_id" binding:"required"`
	Position           string `json:"position,omitempty"`
	Note               string `json:"note,omitempty"`
	MechanicEmployeeID *int64 `json:"mechanic_employee_id,omitempty"`
}

type InstallResponse struct {
	ID             int64     `json:"id"`
	BatchID        int64     `json:"batch_id"`
	BatchNo        string    `json:"batch_no"`
	ProductName    string    `json:"product_name"`
	TireSize       string    `json:"tire_size,omitempty"`
	DOTCode        string    `json:"dot_code,omitempty"`
	VehicleID      *int64    `json:"vehicle_id,omitempty"`
	PlateNumber    string    `json:"plate_number,omitempty"`
	ServiceJobID   *int64    `json:"service_job_id,omitempty"`
	JobNumber      string    `json:"job_number,omitempty"`
	Position       string    `json:"position,omitempty"`
	Note           string    `json:"note,omitempty"`
	MechanicName   string    `json:"mechanic_name,omitempty"`
	InstalledByName string   `json:"installed_by_name,omitempty"`
	InstalledAt    time.Time `json:"installed_at"`
}

// OpenJob is one of today's in-progress/pending jobs the mechanic can attach an
// install to.
type OpenJob struct {
	ID           int64  `json:"id"`
	JobNumber    string `json:"job_number"`
	Status       string `json:"status"`
	VehicleID    *int64 `json:"vehicle_id,omitempty"`
	PlateNumber  string `json:"plate_number,omitempty"`
	Make         string `json:"make,omitempty"`
	Model        string `json:"model,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
}

type Mechanic struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
