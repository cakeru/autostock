package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/employee/dto"
	"github.com/cakeru/autostock/internal/employee/service"
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
	if appErr, ok := err.(*domain.AppError); ok {
		c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
		return
	}
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Employee not found"}})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": msg}})
}

func (h *Handler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context(), branch(c))
	if err != nil {
		fail(c, err, "Failed to list employees")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	emp, err := h.service.Get(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to get employee")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": emp})
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	emp, err := h.service.Create(c.Request.Context(), branch(c), &req)
	if err != nil {
		fail(c, err, "Failed to create employee")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": emp})
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req dto.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	emp, err := h.service.Update(c.Request.Context(), branch(c), id, &req)
	if err != nil {
		fail(c, err, "Failed to update employee")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": emp})
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := h.service.Deactivate(c.Request.Context(), branch(c), id); err != nil {
		fail(c, err, "Failed to delete employee")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"deleted": true}})
}

func (h *Handler) CreateAccount(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req dto.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	emp, err := h.service.CreateAccount(c.Request.Context(), branch(c), id, &req)
	if err != nil {
		fail(c, err, "Failed to create account")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": emp})
}
