package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/supplier/dto"
	"github.com/cakeru/autostock/internal/supplier/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{service: service.NewService(pool)}
}

func branch(c *gin.Context) int64 { v, _ := c.Get("branch_id"); return v.(int64) }

func idParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid id"}})
		return 0, false
	}
	return id, true
}

func fail(c *gin.Context, err error, msg string) {
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Supplier not found"}})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": msg}})
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context(), branch(c))
	if err != nil {
		fail(c, err, "Failed to list suppliers")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	s, err := h.service.Get(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to get supplier")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": s})
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	s, err := h.service.Create(c.Request.Context(), branch(c), &req)
	if err != nil {
		fail(c, err, "Failed to create supplier")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": s})
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req dto.UpdateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	s, err := h.service.Update(c.Request.Context(), branch(c), id, &req)
	if err != nil {
		fail(c, err, "Failed to update supplier")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": s})
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := h.service.Deactivate(c.Request.Context(), branch(c), id); err != nil {
		fail(c, err, "Failed to delete supplier")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"deleted": true}})
}

func (h *Handler) Purchases(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	items, err := h.service.Purchases(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to get purchases")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) Pay(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req dto.PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	s, err := h.service.Pay(c.Request.Context(), branch(c), id, req.Amount)
	if err != nil {
		fail(c, err, "Failed to record payment")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": s})
}
