package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

type SearchResult struct {
	Type   string `json:"type"`
	ID     int64  `json:"id"`
	Label  string `json:"label"`
	Sub    string `json:"sub"`
	URL    string `json:"url"`
}

func (h *Handler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" || len(q) < 2 {
		c.JSON(http.StatusOK, gin.H{"data": []SearchResult{}})
		return
	}

	branchID, _ := c.Get("branch_id")
	like := "%" + q + "%"

	results := make([]SearchResult, 0, 10)

	// Customers (name, phone)
	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, name, COALESCE(phone, '') FROM customers
		WHERE branch_id = $1 AND (name ILIKE $2 OR phone ILIKE $2)
		LIMIT 5`, branchID, like)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var r SearchResult
			var phone string
			if err := rows.Scan(&r.ID, &r.Label, &phone); err == nil {
				r.Type = "customer"
				r.Sub = phone
				r.URL = "/customers/" + formatInt(r.ID)
				results = append(results, r)
			}
		}
	}

	// Vehicles (plate)
	rows2, err := h.pool.Query(c.Request.Context(), `
		SELECT v.id, v.plate_number, COALESCE(c.name, ''), v.customer_id
		FROM vehicles v
		LEFT JOIN customers c ON c.id = v.customer_id
		WHERE v.branch_id = $1 AND v.plate_number ILIKE $2
		LIMIT 5`, branchID, like)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var r SearchResult
			var custName string
			var custID int64
			if err := rows2.Scan(&r.ID, &r.Label, &custName, &custID); err == nil {
				r.Type = "vehicle"
				r.Sub = custName
				r.URL = "/vehicles/" + formatInt(r.ID)
				results = append(results, r)
			}
		}
	}

	// Invoices (number)
	rows3, err := h.pool.Query(c.Request.Context(), `
		SELECT id, invoice_number FROM invoices
		WHERE branch_id = $1 AND invoice_number ILIKE $2
		LIMIT 5`, branchID, like)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var r SearchResult
			if err := rows3.Scan(&r.ID, &r.Label); err == nil {
				r.Type = "invoice"
				r.URL = "/invoices/" + formatInt(r.ID)
				results = append(results, r)
			}
		}
	}

	// Service Jobs (number)
	rows4, err := h.pool.Query(c.Request.Context(), `
		SELECT id, job_number FROM service_jobs
		WHERE branch_id = $1 AND job_number ILIKE $2
		LIMIT 5`, branchID, like)
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var r SearchResult
			if err := rows4.Scan(&r.ID, &r.Label); err == nil {
				r.Type = "job"
				r.URL = "/service-jobs/" + formatInt(r.ID)
				results = append(results, r)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 20)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	if neg {
		digits = append(digits, '-')
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	if len(digits) == 0 {
		return "0"
	}
	return string(digits)
}
