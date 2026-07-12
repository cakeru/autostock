package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cakeru/autostock/internal/auth/dto"
	"github.com/cakeru/autostock/internal/auth/service"
	"github.com/cakeru/autostock/internal/domain"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()},
		})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		fmt.Fprintf(os.Stderr, "LOGIN ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Login failed"},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.service.GetMe(c.Request.Context(), userID.(int64))
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get user"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *Handler) ListUsers(c *gin.Context) {
	branchID, _ := c.Get("branch_id")

	users, err := h.service.ListUsers(c.Request.Context(), branchID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list users"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()},
		})
		return
	}

	branchID, _ := c.Get("branch_id")

	user, err := h.service.CreateUser(c.Request.Context(), &req, branchID.(int64))
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create user"},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid user ID"},
		})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()},
		})
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to update user"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid user ID"},
		})
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to delete user"},
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()},
		})
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), userID.(int64), &req); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			c.JSON(appErr.Status, gin.H{"error": gin.H{"code": appErr.Code, "message": appErr.Message}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to change password"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
