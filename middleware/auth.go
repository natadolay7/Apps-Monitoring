package middleware

import (
	"net/http"
	"strings"

	"api_patroliku_docker/config"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware memvalidasi JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid authorization format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := config.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid or expired token",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// Set user data ke context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("userType", claims.UserType)
		c.Set("branchID", claims.BranchID)

		c.Next()
	}
}
