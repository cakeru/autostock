package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/purchaseorder/dto"
	"github.com/cakeru/autostock/internal/purchaseorder/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{service: service.NewService(pool)}
}

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

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreatePORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	po, err := h.service.Create(c.Request.Context(), branch(c), user(c), &req)
	if err != nil {
		fail(c, err, "Failed to create purchase order")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": po})
}

func (h *Handler) Update(c *gin.Context) {
	branchID, _ := c.Get("branch_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid purchase order id"}})
		return
	}
	var req dto.CreatePORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	po, err := h.service.Update(c.Request.Context(), branchID.(int64), id, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Purchase order or supplier not found"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": po})
}

func (h *Handler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context(), branch(c))
	if err != nil {
		fail(c, err, "Failed to list purchase orders")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	po, err := h.service.Get(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to get purchase order")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": po})
}

func (h *Handler) AddItem(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req dto.AddPOItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	item, err := h.service.AddItem(c.Request.Context(), branch(c), id, &req)
	if err != nil {
		fail(c, err, "Failed to add item")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *Handler) RemoveItem(c *gin.Context) {
	itemID, ok := idParam(c, "item_id")
	if !ok {
		return
	}
	if err := h.service.RemoveItem(c.Request.Context(), branch(c), itemID); err != nil {
		fail(c, err, "Failed to remove item")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) Place(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	po, err := h.service.Place(c.Request.Context(), branch(c), id, user(c))
	if err != nil {
		fail(c, err, "Failed to place purchase order")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": po})
}

func (h *Handler) Cancel(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Cancel(c.Request.Context(), branch(c), id); err != nil {
		fail(c, err, "Failed to cancel purchase order")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"cancelled": true}})
}

func (h *Handler) Receive(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req dto.ReceiveRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	result, err := h.service.Receive(c.Request.Context(), branch(c), id, user(c), &req)
	if err != nil {
		fail(c, err, "Failed to receive purchase order")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
