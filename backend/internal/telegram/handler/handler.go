package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/telegram/dto"
	"github.com/cakeru/autostock/internal/telegram/models"
	"github.com/cakeru/autostock/internal/telegram/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func branch(c *gin.Context) int64 { v, _ := c.Get("branch_id"); return v.(int64) }

func fail(c *gin.Context, status int, code, msg string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
}

func (h *Handler) GetChannels(c *gin.Context) {
	channels, err := h.service.GetChannels(c.Request.Context(), branch(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load channels")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dto.ChannelsResponse{Channels: channels}})
}

func (h *Handler) SaveChannels(c *gin.Context) {
	var req dto.SaveChannelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if err := h.service.SaveChannels(c.Request.Context(), branch(c), req.Channels); err != nil {
		fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save channels")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"saved": true}})
}

func (h *Handler) GetRoutes(c *gin.Context) {
	routes, err := h.service.GetRoutes(c.Request.Context(), branch(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load routes")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dto.RoutesResponse{Routes: routes}})
}

func (h *Handler) SaveRoutes(c *gin.Context) {
	var req dto.SaveRoutesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if err := h.service.SaveRoutes(c.Request.Context(), branch(c), req.Routes); err != nil {
		fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save routes")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"saved": true}})
}

func (h *Handler) TestSend(c *gin.Context) {
	var req dto.TestSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if err := h.service.TestSend(c.Request.Context(), branch(c), req.ChannelID); err != nil {
		fail(c, http.StatusBadGateway, "TELEGRAM_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"sent": true}})
}

// SendDocument relays an uploaded file (an invoice or vehicle-report PDF the
// frontend generated) to the channel routed for the "documents" topic.
func (h *Handler) SendDocument(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "INVALID_REQUEST", "file is required")
		return
	}
	if fileHeader.Size > 20<<20 { // Telegram bot upload cap is 50MB; keep well under
		fail(c, http.StatusBadRequest, "FILE_TOO_LARGE", "File exceeds 20 MB")
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		fail(c, http.StatusBadRequest, "INVALID_REQUEST", "could not read upload")
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not read upload")
		return
	}

	filename := fileHeader.Filename
	if filename == "" {
		filename = "document.pdf"
	}
	caption := c.PostForm("caption")

	if err := h.service.SendDocumentToTopic(c.Request.Context(), branch(c), models.TopicDocuments, filename, data, caption); err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			fail(c, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		fail(c, http.StatusBadGateway, "TELEGRAM_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"sent": true}})
}

func (h *Handler) Trigger(c *gin.Context) {
	var req dto.TriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if err := h.service.TriggerNow(c.Request.Context(), branch(c), req.Topic); err != nil {
		fail(c, http.StatusBadRequest, "TRIGGER_FAILED", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"triggered": true}})
}
