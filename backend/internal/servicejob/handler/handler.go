package handler

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/servicejob/dto"
	"github.com/cakeru/autostock/internal/servicejob/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{service: service.NewService(pool)}
}

func (h *Handler) List(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	var filter dto.ServiceJobFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	jobs, total, err := h.service.List(c.Request.Context(), branchID.(int64), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list jobs"}})
		return
	}

	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}

	c.JSON(http.StatusOK, gin.H{
		"data": jobs,
		"meta": gin.H{"page": filter.Page, "per_page": perPage, "total": total, "total_pages": int(math.Ceil(float64(total) / float64(perPage)))},
	})
}

func (h *Handler) Get(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid job ID"}})
		return
	}

	job, err := h.service.Get(c.Request.Context(), branchID.(int64), id)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get job"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": job})
}

func (h *Handler) Create(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")

	var req dto.CreateServiceJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	job, err := h.service.Create(c.Request.Context(), branchID.(int64), userID.(int64), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create job"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": job})
}

func (h *Handler) Update(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid job ID"}})
		return
	}

	var req dto.UpdateServiceJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	job, err := h.service.Update(c.Request.Context(), branchID.(int64), id, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": job})
}

func (h *Handler) Delete(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid job ID"}})
		return
	}

	if err := h.service.Delete(c.Request.Context(), branchID.(int64), id); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to delete job"}})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) AddItem(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	jobID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid job ID"}})
		return
	}

	var req dto.AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	item, err := h.service.AddItem(c.Request.Context(), branchID.(int64), jobID, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to add item"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *Handler) RemoveItem(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	itemID, err := strconv.ParseInt(c.Param("item_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid item ID"}})
		return
	}

	if err := h.service.RemoveItem(c.Request.Context(), branchID.(int64), itemID); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to remove item"}})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) Complete(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid job ID"}})
		return
	}

	job, err := h.service.Complete(c.Request.Context(), branchID.(int64), id)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": job})
}

func (h *Handler) ApproveQuote(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid job ID"}})
		return
	}

	job, err := h.service.ApproveQuote(c.Request.Context(), branchID.(int64), id, userID.(int64))
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": job})
}
