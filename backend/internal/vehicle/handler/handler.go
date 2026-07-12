package handler

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/storage"
	"github.com/cakeru/autostock/internal/vehicle/dto"
	"github.com/cakeru/autostock/internal/vehicle/service"
)

type Handler struct {
	service *service.Service
	store   storage.Storage
}

func NewHandler(pool *pgxpool.Pool, store storage.Storage) *Handler {
	return &Handler{service: service.NewService(pool), store: store}
}

// maxPhotoBytes/maxPhotoDim are larger than product thumbnails: these are
// alignment printouts and damage photos where a mechanic (or the customer)
// needs to actually read fine detail, not just recognize the item.
const maxPhotoBytes = 10 << 20 // 10 MB upload cap
const maxPhotoDim = 1600

func branch(c *gin.Context) int64 { v, _ := c.Get("branch_id"); return v.(int64) }
func user(c *gin.Context) int64   { v, _ := c.Get("user_id"); return v.(int64) }

func idParam(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
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
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Not found"}})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": msg}})
}

func (h *Handler) GetProfile(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	v, err := h.service.GetProfile(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to get vehicle")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": v})
}

func (h *Handler) GetHistory(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	items, err := h.service.GetHistory(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to get vehicle history")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *Handler) GetTimeline(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	visits, err := h.service.GetTimeline(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to get service timeline")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": visits})
}

func (h *Handler) ListRecords(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	records, err := h.service.ListRecords(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to list records")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": records})
}

func (h *Handler) CreateRecord(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req dto.CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	rec, err := h.service.CreateRecord(c.Request.Context(), branch(c), id, user(c), &req)
	if err != nil {
		fail(c, err, "Failed to create record")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": rec})
}

func (h *Handler) DeleteRecord(c *gin.Context) {
	id, ok := idParam(c, "record_id")
	if !ok {
		return
	}
	urls, err := h.service.DeleteRecord(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to delete record")
		return
	}
	for _, u := range urls {
		_ = h.store.Delete(c.Request.Context(), u)
	}
	c.JSON(http.StatusNoContent, nil)
}

// processUploadedPhoto reads the "photo" form field, resizes/orients/encodes it
// to JPEG, stores it under keyPrefix, and returns the stored URL. On any failure
// it writes the error response itself and returns ok=false. Shared by the
// vehicle-record and wheel-service photo endpoints.
func (h *Handler) processUploadedPhoto(c *gin.Context, keyPrefix string, id int64) (string, bool) {
	fileHeader, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Expected a 'photo' file field"}})
		return "", false
	}
	if fileHeader.Size > maxPhotoBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"code": "FILE_TOO_LARGE", "message": "Photo must be 10 MB or smaller"}})
		return "", false
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Could not read upload"}})
		return "", false
	}
	defer file.Close()

	img, err := imaging.Decode(file, imaging.AutoOrientation(true))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_IMAGE", "message": "Unsupported or corrupt image file"}})
		return "", false
	}
	img = imaging.Fit(img, maxPhotoDim, maxPhotoDim, imaging.Lanczos)
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to process image"}})
		return "", false
	}

	key := fmt.Sprintf("%s/%d-%d.jpg", keyPrefix, id, time.Now().UnixNano())
	url, err := h.store.Save(c.Request.Context(), key, &buf, "image/jpeg")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "STORAGE_ERROR", "message": "Failed to store image"}})
		return "", false
	}
	return url, true
}

func (h *Handler) AddPhoto(c *gin.Context) {
	recordID, ok := idParam(c, "record_id")
	if !ok {
		return
	}
	url, ok := h.processUploadedPhoto(c, "vehicle-records", recordID)
	if !ok {
		return
	}
	photo, err := h.service.AddPhoto(c.Request.Context(), branch(c), recordID, url)
	if err != nil {
		_ = h.store.Delete(c.Request.Context(), url)
		fail(c, err, "Failed to save photo reference")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": photo})
}

func (h *Handler) DeletePhoto(c *gin.Context) {
	id, ok := idParam(c, "photo_id")
	if !ok {
		return
	}
	url, err := h.service.DeletePhoto(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to delete photo")
		return
	}
	_ = h.store.Delete(c.Request.Context(), url)
	c.JSON(http.StatusNoContent, nil)
}
