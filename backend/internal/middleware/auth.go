package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Claims struct {
	UserID      int64    `json:"user_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	BranchID    int64    `json:"branch_id"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func AuthMiddleware(jwtSecret string, pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Missing authorization header"},
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid authorization format"},
			})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Invalid or expired token"},
			})
			return
		}

		var isActive bool
		err = pool.QueryRow(c.Request.Context(), `SELECT is_active FROM users WHERE id = $1`, claims.UserID).Scan(&isActive)
		if err != nil || !isActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "User is deactivated"},
			})
			return
		}

		var dbPermissions []string
		rows, err := pool.Query(c.Request.Context(), `SELECT unnest(permissions) FROM users WHERE id = $1`, claims.UserID)
		if err == nil {
			dbPermissions = []string{}
			for rows.Next() {
				var p string
				if err := rows.Scan(&p); err == nil {
					dbPermissions = append(dbPermissions, p)
				}
			}
			rows.Close()
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("branch_id", claims.BranchID)
		c.Set("permissions", dbPermissions)
		c.Next()
	}
}
