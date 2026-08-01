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
	"github.com/cakeru/autostock/internal/invoice/dto"
	"github.com/cakeru/autostock/internal/invoice/service"
	"github.com/cakeru/autostock/internal/storage"
)

type Handler struct {
	service *service.Service
	store   storage.Storage
}

func NewHandler(pool *pgxpool.Pool, store storage.Storage) *Handler {
	return &Handler{service: service.NewService(pool), store: store}
}

const maxProofBytes = 10 << 20 // 10 MB upload cap
const maxProofDim = 1600

func (h *Handler) List(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	var filter dto.InvoiceFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	invoices, total, err := h.service.List(c.Request.Context(), branchID.(int64), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list invoices"}})
		return
	}

	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}

	c.JSON(http.StatusOK, gin.H{
		"data": invoices,
		"meta": gin.H{"page": filter.Page, "per_page": perPage, "total": total, "total_pages": int(math.Ceil(float64(total) / float64(perPage)))},
	})
}

func (h *Handler) Get(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid invoice ID"},
		})
		return
	}

	invoice, err := h.service.Get(c.Request.Context(), branchID.(int64), id)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invoice})
}

func (h *Handler) Create(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")

	var req dto.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	invoice, err := h.service.Create(c.Request.Context(), branchID.(int64), userID.(int64), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": invoice})
}

func (h *Handler) CreateFromJob(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")
	jobID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid job ID"}})
		return
	}

	var req dto.CreateInvoiceFromJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	invoice, err := h.service.CreateFromServiceJob(c.Request.Context(), branchID.(int64), userID.(int64), jobID, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": invoice})
}

func (h *Handler) Update(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid invoice ID"}})
		return
	}

	var req dto.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	invoice, err := h.service.Update(c.Request.Context(), branchID.(int64), id, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to update invoice"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invoice})
}

func (h *Handler) Void(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid invoice ID"}})
		return
	}

	var req dto.VoidInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	invoice, err := h.service.Void(c.Request.Context(), branchID.(int64), id, userID.(int64), req.Reason)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invoice})
}

func (h *Handler) RecordPayment(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid invoice ID"}})
		return
	}

	var req dto.RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	payment, err := h.service.RecordPayment(c.Request.Context(), branchID.(int64), id, userID.(int64), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": payment})
}

func (h *Handler) ListPayments(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid invoice ID"}})
		return
	}

	payments, err := h.service.GetPayments(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list payments"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": payments})
}

func (h *Handler) PDF(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "PDF generation not yet implemented"})
}

func (h *Handler) AddItem(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")
	invoiceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid invoice ID"}})
		return
	}

	var req dto.InvoiceItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	item, err := h.service.AddItem(c.Request.Context(), branchID.(int64), userID.(int64), invoiceID, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *Handler) UpdateItem(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")
	invoiceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid invoice ID"}})
		return
	}
	itemID, err := strconv.ParseInt(c.Param("item_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid item ID"}})
		return
	}

	var req dto.UpdateInvoiceItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}

	item, err := h.service.UpdateItem(c.Request.Context(), branchID.(int64), userID.(int64), invoiceID, itemID, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *Handler) RemoveItem(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	userID, _ := c.Get("user_id")
	itemID, err := strconv.ParseInt(c.Param("item_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid item ID"}})
		return
	}

	if err := h.service.RemoveItem(c.Request.Context(), branchID.(int64), userID.(int64), itemID); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item removed"})
}

// UploadPaymentProof accepts a bank-transfer confirmation photo ("photo" form
// field), stores it like other uploads, and attaches it to the payment.
func (h *Handler) UploadPaymentProof(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	invoiceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid invoice ID"}})
		return
	}
	paymentID, err := strconv.ParseInt(c.Param("payment_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid payment ID"}})
		return
	}

	fileHeader, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Expected a 'photo' file field"}})
		return
	}
	if fileHeader.Size > maxProofBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"code": "FILE_TOO_LARGE", "message": "Photo must be 10 MB or smaller"}})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": "Could not read upload"}})
		return
	}
	defer file.Close()

	img, err := imaging.Decode(file, imaging.AutoOrientation(true))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_IMAGE", "message": "Unsupported or corrupt image file"}})
		return
	}
	img = imaging.Fit(img, maxProofDim, maxProofDim, imaging.Lanczos)
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to process image"}})
		return
	}

	key := fmt.Sprintf("payment-proofs/%d-%d.jpg", invoiceID, time.Now().UnixNano())
	url, err := h.store.Save(c.Request.Context(), key, &buf, "image/jpeg")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "STORAGE_ERROR", "message": "Failed to store image"}})
		return
	}

	if err := h.service.SetPaymentProof(c.Request.Context(), branchID.(int64), paymentID, url); err != nil {
		_ = h.store.Delete(c.Request.Context(), url)
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to save payment proof"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"payment_id": paymentID, "proof_url": url}})
}
