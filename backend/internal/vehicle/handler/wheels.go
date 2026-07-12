package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cakeru/autostock/internal/vehicle/dto"
)

func (h *Handler) ListWheelServices(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	services, err := h.service.ListWheelServices(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to list wheel services")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": services})
}

func (h *Handler) CreateWheelService(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req dto.CreateWheelServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	svc, err := h.service.CreateWheelService(c.Request.Context(), branch(c), id, user(c), &req)
	if err != nil {
		fail(c, err, "Failed to create wheel service")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": svc})
}

func (h *Handler) DeleteWheelService(c *gin.Context) {
	id, ok := idParam(c, "service_id")
	if !ok {
		return
	}
	urls, err := h.service.DeleteWheelService(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to delete wheel service")
		return
	}
	for _, u := range urls {
		_ = h.store.Delete(c.Request.Context(), u)
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) AddWheelServicePhoto(c *gin.Context) {
	serviceID, ok := idParam(c, "service_id")
	if !ok {
		return
	}
	url, ok := h.processUploadedPhoto(c, "wheel-services", serviceID)
	if !ok {
		return
	}
	photo, err := h.service.AddWheelServicePhoto(c.Request.Context(), branch(c), serviceID, url)
	if err != nil {
		_ = h.store.Delete(c.Request.Context(), url)
		fail(c, err, "Failed to save photo reference")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": photo})
}

func (h *Handler) RecentTireOptions(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	opts, err := h.service.RecentTireOptions(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to list tire options")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": opts})
}

// ---------------------------------------------------------------------------
// Parts log
// ---------------------------------------------------------------------------

func (h *Handler) ListParts(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	parts, err := h.service.ListParts(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to list parts")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": parts})
}

func (h *Handler) CreatePart(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req dto.CreatePartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	part, err := h.service.CreatePart(c.Request.Context(), branch(c), id, user(c), &req)
	if err != nil {
		fail(c, err, "Failed to log part")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": part})
}

func (h *Handler) DeletePart(c *gin.Context) {
	id, ok := idParam(c, "part_id")
	if !ok {
		return
	}
	if err := h.service.DeletePart(c.Request.Context(), branch(c), id); err != nil {
		fail(c, err, "Failed to delete part")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ---------------------------------------------------------------------------
// DVI part statuses
// ---------------------------------------------------------------------------

func (h *Handler) ListPartStatuses(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	statuses, err := h.service.ListPartStatuses(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to list part statuses")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": statuses})
}

func (h *Handler) SetPartStatus(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req dto.SetPartStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	if err := h.service.SetPartStatus(c.Request.Context(), branch(c), id, user(c), &req); err != nil {
		fail(c, err, "Failed to set part status")
		return
	}
	statuses, err := h.service.ListPartStatuses(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to reload part statuses")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": statuses})
}
