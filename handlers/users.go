package handlers

import (
	"net/http"

	"GoCall_api/db"

	"github.com/gin-gonic/gin"
)

// GetUserID returns the authenticated user's UUID
func GetUserID(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var user db.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"userID": user.UserID})
}