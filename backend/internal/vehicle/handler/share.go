package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cakeru/autostock/internal/vehicle/dto"
)

func (h *Handler) EnsureShareLink(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	token, err := h.service.EnsureShareLink(c.Request.Context(), branch(c), id)
	if err != nil {
		fail(c, err, "Failed to create share link")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dto.ShareLinkResponse{Token: token}})
}

func (h *Handler) RevokeShareLink(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.RevokeShareLink(c.Request.Context(), branch(c), id); err != nil {
		fail(c, err, "Failed to revoke share link")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// PublicReport is the one unauthenticated vehicle endpoint: the share token in
// the URL is the entire authorization.
func (h *Handler) PublicReport(c *gin.Context) {
	token := c.Param("token")
	report, err := h.service.GetPublicReport(c.Request.Context(), token)
	if err != nil {
		fail(c, err, "Report not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}
