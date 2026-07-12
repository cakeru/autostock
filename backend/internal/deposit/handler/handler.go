package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/deposit/dto"
	"github.com/cakeru/autostock/internal/deposit/service"
	"github.com/cakeru/autostock/internal/domain"
)

type Handler struct {
	service *service.Service
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{service: service.NewService(pool)}
}

func branch(c *gin.Context) int64 { v, _ := c.Get("branch_id"); return v.(int64) }
func user(c *gin.Context) int64   { v, _ := c.Get("user_id"); return v.(int64) }

func respond(c *gin.Context, err error, status int, data interface{}) {
	if err == nil {
		c.JSON(status, gin.H{"data": data})
		return
	}
	if appErr, ok := err.(*domain.AppError); ok {
		c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
		return
	}
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Deposit not found"}})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Request failed"}})
}

func (h *Handler) List(c *gin.Context) {
	customerID, _ := strconv.ParseInt(c.Query("customer_id"), 10, 64)
	items, err := h.service.List(c.Request.Context(), branch(c), customerID, c.Query("status"))
	respond(c, err, http.StatusOK, items)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	d, err := h.service.Create(c.Request.Context(), branch(c), user(c), &req)
	respond(c, err, http.StatusCreated, d)
}

func (h *Handler) Apply(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req dto.ApplyDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	d, err := h.service.Apply(c.Request.Context(), branch(c), user(c), id, req.InvoiceID)
	respond(c, err, http.StatusOK, d)
}

func (h *Handler) Refund(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	d, err := h.service.Refund(c.Request.Context(), branch(c), id)
	respond(c, err, http.StatusOK, d)
}
