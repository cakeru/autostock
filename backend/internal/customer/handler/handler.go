package handler

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/customer/dto"
	"github.com/cakeru/autostock/internal/customer/service"
	"github.com/cakeru/autostock/internal/domain"
)

type Handler struct {
	service *service.Service
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{service: service.NewService(pool)}
}

func (h *Handler) List(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	var filter dto.CustomerFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()},
		})
		return
	}

	customers, total, err := h.service.List(c.Request.Context(), branchID.(int64), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list customers"},
		})
		return
	}

	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}

	c.JSON(http.StatusOK, gin.H{
		"data": customers,
		"meta": gin.H{
			"page":        filter.Page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": int(math.Ceil(float64(total) / float64(perPage))),
		},
	})
}

func (h *Handler) Get(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid customer ID"},
		})
		return
	}

	cust, vehicles, err := h.service.Get(c.Request.Context(), branchID.(int64), id)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get customer"},
		})
		return
	}

	stats, err := h.service.GetStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get customer stats"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":             cust.ID,
			"name":           cust.Name,
			"phone":          cust.Phone,
			"email":          cust.Email,
			"address":        cust.Address,
			"notes":          cust.Notes,
			"customer_since": cust.CustomerSince,
			"is_active":      cust.IsActive,
			"created_at":     cust.CreatedAt,
			"updated_at":     cust.UpdatedAt,
			"vehicles":       vehicles,
			"total_spent":    stats.TotalSpent,
			"visit_count":    stats.VisitCount,
			"last_visit":     stats.LastVisit,
			"outstanding":    stats.Outstanding,
		},
	})
}

func (h *Handler) Create(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	var req dto.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	cust, err := h.service.Create(c.Request.Context(), branchID.(int64), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create customer"},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": cust})
}

func (h *Handler) Update(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid customer ID"},
		})
		return
	}

	var req dto.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	cust, err := h.service.Update(c.Request.Context(), branchID.(int64), id, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to update customer"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cust})
}

func (h *Handler) Delete(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid customer ID"},
		})
		return
	}

	if err := h.service.Delete(c.Request.Context(), branchID.(int64), id); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to delete customer"},
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) History(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid customer ID"},
		})
		return
	}

	items, err := h.service.GetActivity(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get customer activity"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) ListVehicles(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid customer ID"},
		})
		return
	}

	vehicles, err := h.service.ListVehicles(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list vehicles"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vehicles})
}

func (h *Handler) CreateVehicle(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid customer ID"},
		})
		return
	}

	var req dto.CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	vehicle, err := h.service.CreateVehicle(c.Request.Context(), customerID, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create vehicle"},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": vehicle})
}

func (h *Handler) UpdateVehicle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid vehicle ID"},
		})
		return
	}

	var req dto.UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	vehicle, err := h.service.UpdateVehicle(c.Request.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to update vehicle"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vehicle})
}

func (h *Handler) DeleteVehicle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid vehicle ID"},
		})
		return
	}

	if err := h.service.DeleteVehicle(c.Request.Context(), id); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to delete vehicle"},
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
