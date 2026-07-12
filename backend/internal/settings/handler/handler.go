package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cakeru/autostock/internal/settings/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetSettings(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	settings, err := h.service.GetSettings(c.Request.Context(), branchID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get settings"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func (h *Handler) UpdateSetting(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()},
		})
		return
	}

	if err := h.service.UpdateSetting(c.Request.Context(), branchID.(int64), req.Key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to update setting"},
		})
		return
	}

	settings, err := h.service.GetSettings(c.Request.Context(), branchID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get settings"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func (h *Handler) GetExchangeRate(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	rate, err := h.service.GetExchangeRate(c.Request.Context(), branchID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get exchange rate"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rate})
}

func (h *Handler) UpdateExchangeRate(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	var req struct {
		Rate float64 `json:"rate" binding:"required,min=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()},
		})
		return
	}

	value := strconv.FormatFloat(req.Rate, 'f', 0, 64)
	if err := h.service.UpdateSetting(c.Request.Context(), branchID.(int64), "exchange_rate_usd_khr", value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to update exchange rate"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"rate": req.Rate}})
}
