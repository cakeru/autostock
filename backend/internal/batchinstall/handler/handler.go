package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/batchinstall/dto"
	"github.com/cakeru/autostock/internal/batchinstall/service"
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

func fail(c *gin.Context, err error, msg string) {
	if appErr, ok := err.(*domain.AppError); ok {
		c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
		return
	}
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Not found"}})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": msg}})
}

// Resolve turns a scanned QR payload into batch details for confirmation.
func (h *Handler) Resolve(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "code is required"}})
		return
	}
	info, err := h.service.ResolveCode(c.Request.Context(), branch(c), code)
	if err != nil {
		fail(c, err, "Failed to resolve batch")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": info})
}

func (h *Handler) Record(c *gin.Context) {
	var req dto.RecordInstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	install, err := h.service.RecordInstall(c.Request.Context(), branch(c), user(c), &req)
	if err != nil {
		fail(c, err, "Failed to record install")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": install})
}

func (h *Handler) OpenJobs(c *gin.Context) {
	jobs, err := h.service.OpenJobs(c.Request.Context(), branch(c))
	if err != nil {
		fail(c, err, "Failed to list open jobs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": jobs})
}

func (h *Handler) Mechanics(c *gin.Context) {
	mechanics, err := h.service.Mechanics(c.Request.Context(), branch(c))
	if err != nil {
		fail(c, err, "Failed to list mechanics")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": mechanics})
}
