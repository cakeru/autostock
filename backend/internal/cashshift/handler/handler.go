package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/cashshift/dto"
	"github.com/cakeru/autostock/internal/cashshift/service"
	"github.com/cakeru/autostock/internal/domain"
)

type Handler struct {
	service *service.Service
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{service: service.NewService(pool)}
}

func (h *Handler) Current(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	shift, err := h.service.GetCurrent(c.Request.Context(), branchID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get current shift"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": shift}) // data is null when no drawer is open
}

func (h *Handler) Open(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")

	var req dto.OpenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	shift, err := h.service.Open(c.Request.Context(), branchID.(int64), userID.(int64), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to open drawer"}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": shift})
}

func (h *Handler) Close(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")

	var req dto.CloseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	shift, err := h.service.Close(c.Request.Context(), branchID.(int64), userID.(int64), &req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "No open drawer to close"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to close drawer"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": shift})
}

func (h *Handler) List(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	shifts, err := h.service.List(c.Request.Context(), branchID.(int64), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list shifts"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": shifts})
}
