package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cakeru/autostock/internal/vehicle/dto"
)

func (h *Handler) ListGalleryPhotos(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	photos, err := h.service.ListGalleryPhotos(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to list photos")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": photos})
}

func (h *Handler) AddGalleryPhoto(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	url, ok := h.processUploadedPhoto(c, "vehicle-photos", id)
	if !ok {
		return
	}
	// Optional taken_at (YYYY-MM-DD) dates the photo to the visit it documents.
	var takenAt *time.Time
	if v := c.PostForm("taken_at"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			takenAt = &t
		}
	}
	photo, err := h.service.AddGalleryPhoto(c.Request.Context(), branch(c), id, user(c), url, takenAt)
	if err != nil {
		_ = h.store.Delete(c.Request.Context(), url)
		fail(c, err, "Failed to save photo")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": photo})
}

func (h *Handler) UpdateGalleryPhoto(c *gin.Context) {
	id, ok := idParam(c, "photo_id")
	if !ok {
		return
	}
	var req dto.UpdateGalleryPhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	photo, err := h.service.UpdateGalleryPhoto(c.Request.Context(), branch(c), id, &req)
	if err != nil {
		fail(c, err, "Failed to update photo")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": photo})
}

func (h *Handler) DeleteGalleryPhoto(c *gin.Context) {
	id, ok := idParam(c, "photo_id")
	if !ok {
		return
	}
	url, err := h.service.DeleteGalleryPhoto(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to delete photo")
		return
	}
	_ = h.store.Delete(c.Request.Context(), url)
	c.JSON(http.StatusNoContent, nil)
}
