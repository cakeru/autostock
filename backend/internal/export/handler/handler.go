package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/export/service"
)

type Handler struct {
	service   *service.Service
	backupDir string
}

func NewHandler(pool *pgxpool.Pool, backupDir string) *Handler {
	return &Handler{service: service.NewService(pool), backupDir: backupDir}
}

func streamCSV(c *gin.Context, filename string, write func(c *gin.Context) error) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if err := write(c); err != nil {
		// Headers are already sent; nothing useful to return beyond the log.
		_ = c.Error(err)
	}
}

func (h *Handler) ExportInvoices(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	streamCSV(c, "invoices.csv", func(c *gin.Context) error {
		return h.service.ExportInvoices(c.Request.Context(), branchID.(int64), c.Writer)
	})
}

func (h *Handler) ExportCustomers(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	streamCSV(c, "customers.csv", func(c *gin.Context) error {
		return h.service.ExportCustomers(c.Request.Context(), branchID.(int64), c.Writer)
	})
}

func (h *Handler) ExportProducts(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	streamCSV(c, "products.csv", func(c *gin.Context) error {
		return h.service.ExportProducts(c.Request.Context(), branchID.(int64), c.Writer)
	})
}

// DownloadLatestBackup streams the most recent nightly database dump.
func (h *Handler) DownloadLatestBackup(c *gin.Context) {
	entries, err := os.ReadDir(h.backupDir)
	if err != nil || len(entries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NO_BACKUP", "message": "No backups yet — the nightly job writes its first dump within 24h"}})
		return
	}

	var dumps []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "autostock-") && strings.HasSuffix(e.Name(), ".sql.gz") {
			dumps = append(dumps, e.Name())
		}
	}
	if len(dumps) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NO_BACKUP", "message": "No backups yet — the nightly job writes its first dump within 24h"}})
		return
	}
	sort.Strings(dumps)
	latest := dumps[len(dumps)-1]

	c.Header("Content-Type", "application/gzip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", latest))
	c.File(filepath.Join(h.backupDir, latest))
}
