package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUserID returns the authenticated user's ID
func GetUserID(c *gin.Context) {
    userID, _ := c.Get("user_id")
    c.JSON(http.StatusOK, gin.H{"user_id": userID})
}