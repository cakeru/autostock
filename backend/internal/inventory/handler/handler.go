package handler

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/inventory/dto"
	"github.com/cakeru/autostock/internal/inventory/service"
	"github.com/cakeru/autostock/internal/storage"
)

type Handler struct {
	service *service.Service
	store   storage.Storage
}

func NewHandler(pool *pgxpool.Pool, store storage.Storage) *Handler {
	return &Handler{service: service.NewService(pool), store: store}
}

const maxImageBytes = 8 << 20 // 8 MB upload cap
const maxImportBytes = 5 << 20 // 5 MB CSV upload cap

func (h *Handler) List(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	var filter dto.ProductFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()},
		})
		return
	}

	products, total, err := h.service.List(c.Request.Context(), branchID.(int64), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list products"},
		})
		return
	}

	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
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
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid product ID"},
		})
		return
	}

	product, err := h.service.Get(c.Request.Context(), branchID.(int64), id)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get product"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *Handler) Create(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	userID, _ := c.Get("user_id")
	product, err := h.service.Create(c.Request.Context(), branchID.(int64), userID.(int64), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create product"},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": product})
}

func (h *Handler) Update(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid product ID"},
		})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	product, err := h.service.Update(c.Request.Context(), branchID.(int64), id, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to update product"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *Handler) Delete(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid product ID"},
		})
		return
	}

	if err := h.service.Delete(c.Request.Context(), branchID.(int64), id); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to delete product"},
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) UploadImage(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid product ID"}})
		return
	}

	// Confirm the product exists in this branch and grab any existing image.
	current, err := h.service.Get(c.Request.Context(), branchID.(int64), id)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to load product"}})
		return
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Expected an 'image' file field"}})
		return
	}
	if fileHeader.Size > maxImageBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"code": "FILE_TOO_LARGE", "message": "Image must be 8 MB or smaller"}})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Could not read upload"}})
		return
	}
	defer file.Close()

	// Decode (honoring EXIF orientation from phone cameras), downscale to fit
	// 800px, and re-encode to JPEG so the stored file is small and uniform.
	img, err := imaging.Decode(file, imaging.AutoOrientation(true))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_IMAGE", "message": "Unsupported or corrupt image file"}})
		return
	}
	img = imaging.Fit(img, 800, 800, imaging.Lanczos)
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(82)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to process image"}})
		return
	}

	key := fmt.Sprintf("products/%d-%d.jpg", id, time.Now().UnixNano())
	url, err := h.store.Save(c.Request.Context(), key, &buf, "image/jpeg")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "STORAGE_ERROR", "message": "Failed to store image"}})
		return
	}

	product, err := h.service.SetImage(c.Request.Context(), branchID.(int64), id, url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to save image reference"}})
		return
	}

	// Best-effort cleanup of the file we just replaced (no-op for external URLs).
	if current.ImageURL != "" && current.ImageURL != url {
		_ = h.store.Delete(c.Request.Context(), current.ImageURL)
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *Handler) DeleteImage(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid product ID"}})
		return
	}

	current, err := h.service.Get(c.Request.Context(), branchID.(int64), id)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to load product"}})
		return
	}

	product, err := h.service.SetImage(c.Request.Context(), branchID.(int64), id, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to clear image"}})
		return
	}
	if current.ImageURL != "" {
		_ = h.store.Delete(c.Request.Context(), current.ImageURL)
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *Handler) Movements(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid product ID"}})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	movements, total, err := h.service.ListMovements(c.Request.Context(), branchID.(int64), id, page, perPage)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list movements"}})
		return
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	c.JSON(http.StatusOK, gin.H{
		"data": movements,
		"meta": gin.H{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": int(math.Ceil(float64(total) / float64(perPage))),
		},
	})
}

func (h *Handler) Batches(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid product ID"}})
		return
	}
	batches, err := h.service.ListBatches(c.Request.Context(), branchID.(int64), id)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list batches"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": batches})
}

func (h *Handler) BatchConsumers(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	batchID, err := strconv.ParseInt(c.Param("batch_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid batch ID"}})
		return
	}
	consumers, err := h.service.BatchConsumers(c.Request.Context(), branchID.(int64), batchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list consumers"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": consumers})
}

func (h *Handler) LowStock(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	products, err := h.service.GetLowStock(c.Request.Context(), branchID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get low stock products"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h *Handler) ReceiveStock(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid product ID"}})
		return
	}

	var req dto.ReceiveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	userID, _ := c.Get("user_id")
	product, err := h.service.ReceiveStock(c.Request.Context(), branchID.(int64), id, userID.(int64), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *Handler) Import(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Expected a 'file' field with a CSV upload"}})
		return
	}
	if fileHeader.Size > maxImportBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"code": "FILE_TOO_LARGE", "message": "CSV must be 5 MB or smaller"}})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Could not read upload"}})
		return
	}
	defer file.Close()

	result, err := h.service.ImportCSV(c.Request.Context(), branchID.(int64), userID.(int64), file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "IMPORT_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *Handler) AdjustStock(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid product ID"}})
		return
	}

	var req dto.AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	userID, _ := c.Get("user_id")
	product, err := h.service.AdjustStock(c.Request.Context(), branchID.(int64), id, userID.(int64), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}
