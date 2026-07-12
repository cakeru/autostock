package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/stocktake/dto"
	"github.com/cakeru/autostock/internal/stocktake/service"
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
	var req dto.CreateStocktakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	st, err := h.service.Create(c.Request.Context(), branch(c), user(c), &req)
	if err != nil {
		fail(c, err, "Failed to create stocktake")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": st})
}

func (h *Handler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context(), branch(c))
	if err != nil {
		fail(c, err, "Failed to list stocktakes")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	st, err := h.service.Get(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to get stocktake")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": st})
}

func (h *Handler) AddItem(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req dto.AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	item, err := h.service.AddItem(c.Request.Context(), branch(c), id, req.ProductID)
	if err != nil {
		fail(c, err, "Failed to add item")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *Handler) SetCount(c *gin.Context) {
	itemID, ok := idParam(c, "item_id")
	if !ok {
		return
	}
	var req dto.SetCountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	item, err := h.service.SetCount(c.Request.Context(), branch(c), itemID, user(c), *req.CountedQty)
	if err != nil {
		fail(c, err, "Failed to record count")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
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

func (h *Handler) Cancel(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Cancel(c.Request.Context(), branch(c), id); err != nil {
		fail(c, err, "Failed to cancel stocktake")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"cancelled": true}})
}

func (h *Handler) Finalize(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	result, err := h.service.Finalize(c.Request.Context(), branch(c), id, user(c))
	if err != nil {
		fail(c, err, "Failed to finalize stocktake")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
