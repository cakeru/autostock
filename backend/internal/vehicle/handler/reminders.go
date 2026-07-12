package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cakeru/autostock/internal/vehicle/dto"
)

func (h *Handler) ListServiceEvents(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	events, err := h.service.ListServiceEvents(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to list service events")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": events})
}

func (h *Handler) CreateServiceEvent(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req dto.CreateServiceEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	e, err := h.service.CreateServiceEvent(c.Request.Context(), branch(c), id, user(c), &req)
	if err != nil {
		fail(c, err, "Failed to create service event")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": e})
}

func (h *Handler) DeleteServiceEvent(c *gin.Context) {
	id, ok := idParam(c, "event_id")
	if !ok {
		return
	}
	if err := h.service.DeleteServiceEvent(c.Request.Context(), branch(c), id); err != nil {
		fail(c, err, "Failed to delete service event")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) UpdateVehicleIntervals(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateVehicleIntervalsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	if err := h.service.UpdateVehicleIntervals(c.Request.Context(), branch(c), id, &req); err != nil {
		fail(c, err, "Failed to update intervals")
		return
	}
	v, err := h.service.GetProfile(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to reload vehicle")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": v})
}

func (h *Handler) ListDueForService(c *gin.Context) {
	// horizon_days>0 also returns on-track items whose due date falls within the
	// window (for the calendar); default 0 = overdue + due-soon only (call list).
	horizon := 0
	if v := c.Query("horizon_days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > 365 {
				n = 365
			}
			horizon = n
		}
	}
	items, err := h.service.ListDueForService(c.Request.Context(), branch(c), horizon)
	if err != nil {
		fail(c, err, "Failed to list vehicles due for service")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) GetIntervalSettings(c *gin.Context) {
	s, err := h.service.GetIntervalSettings(c.Request.Context(), branch(c))
	if err != nil {
		fail(c, err, "Failed to load interval settings")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": s})
}

func (h *Handler) UpdateIntervalSettings(c *gin.Context) {
	var req dto.UpdateIntervalSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	if err := h.service.UpdateIntervalSettings(c.Request.Context(), branch(c), &req); err != nil {
		fail(c, err, "Failed to update interval settings")
		return
	}
	s, err := h.service.GetIntervalSettings(c.Request.Context(), branch(c))
	if err != nil {
		fail(c, err, "Failed to reload interval settings")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": s})
}
