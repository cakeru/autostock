package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/analytics/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{service: service.NewService(pool)}
}

func monthStart() string {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
}

func (h *Handler) Sales(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	from := c.DefaultQuery("from", monthStart())
	to := c.DefaultQuery("to", time.Now().Format("2006-01-02"))
	gran := c.DefaultQuery("granularity", "day")

	res, err := h.service.GetSales(c.Request.Context(), branchID.(int64), from, to, gran)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get sales analytics"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (h *Handler) Receivables(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	res, err := h.service.GetReceivables(c.Request.Context(), branchID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get receivables"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (h *Handler) Inventory(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))
	res, err := h.service.GetInventory(c.Request.Context(), branchID.(int64), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get inventory analytics"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (h *Handler) Customers(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))
	res, err := h.service.GetCustomers(c.Request.Context(), branchID.(int64), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get customer analytics"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (h *Handler) Technicians(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	from := c.DefaultQuery("from", monthStart())
	to := c.DefaultQuery("to", time.Now().Format("2006-01-02"))
	res, err := h.service.GetTechnicians(c.Request.Context(), branchID.(int64), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get technician report"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (h *Handler) PnL(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	from := c.DefaultQuery("from", monthStart())
	to := c.DefaultQuery("to", time.Now().Format("2006-01-02"))
	res, err := h.service.GetPnL(c.Request.Context(), branchID.(int64), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get P&L"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": res})
}
