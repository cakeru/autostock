package handler

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/audit/dto"
	"github.com/cakeru/autostock/internal/audit/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{service: service.NewService(pool)}
}

func (h *Handler) List(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	var filter dto.AuditFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	items, total, err := h.service.List(c.Request.Context(), branchID.(int64), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list audit logs"}})
		return
	}

	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 25
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": gin.H{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": int(math.Ceil(float64(total) / float64(perPage))),
		},
	})
}
