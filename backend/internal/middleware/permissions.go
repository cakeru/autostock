package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role == "admin" {
			c.Next()
			return
		}

		perms, exists := c.Get("permissions")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "FORBIDDEN", "message": "No permissions found"},
			})
			return
		}

		permissions, ok := perms.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "FORBIDDEN", "message": "Invalid permissions format"},
			})
			return
		}

		for _, p := range permissions {
			if p == requiredPermission {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": gin.H{"code": "FORBIDDEN", "message": "Insufficient permissions"},
		})
	}
}
