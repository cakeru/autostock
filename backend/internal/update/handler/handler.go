package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler proxies the Settings "Update now" button to the updater container,
// which holds the Docker socket and the repo checkout.
type Handler struct {
	updaterURL string // e.g. http://updater:8081
}

func NewHandler(updaterURL string) *Handler {
	return &Handler{updaterURL: updaterURL}
}

func (h *Handler) Deploy(c *gin.Context) {
	if h.updaterURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "UPDATE_DISABLED", "message": "Auto-update is not configured on this server"}})
		return
	}
	resp, err := http.Post(h.updaterURL+"/deploy", "application/json", nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"code": "UPDATE_UNAVAILABLE", "message": "The updater agent is not reachable"}})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, "application/json", body)
}

func (h *Handler) Status(c *gin.Context) {
	if h.updaterURL == "" {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "disabled", "log_tail": ""}})
		return
	}
	resp, err := http.Get(h.updaterURL + "/status")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "unavailable", "log_tail": ""}})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	c.Data(http.StatusOK, "application/json", body)
}
