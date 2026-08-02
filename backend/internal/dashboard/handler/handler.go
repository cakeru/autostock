package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/dashboard/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{service: service.NewService(pool)}
}

func (h *Handler) Summary(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	summary, err := h.service.GetSummary(c.Request.Context(), branchID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get dashboard summary"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

func (h *Handler) DailyRevenue(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	items, err := h.service.GetDailyRevenue(c.Request.Context(), branchID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get daily revenue"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) Profit(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	now := time.Now()
	from := c.DefaultQuery("from", now.AddDate(0, 0, -1*(now.Day()-1)).Format("2006-01-02"))
	to := c.DefaultQuery("to", now.Format("2006-01-02"))

	profit, err := h.service.GetProfit(c.Request.Context(), branchID.(int64), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get profit report"},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": profit})
}

func (h *Handler) DayClose(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	date := c.DefaultQuery("date", "")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	summary, err := h.service.GetDayClose(c.Request.Context(), branchID.(int64), date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get day close summary"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}
