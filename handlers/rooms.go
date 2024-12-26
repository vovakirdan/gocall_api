package handlers

import (
	"GoCall_api/db"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetRooms retrieves the list of rooms created by the authenticated user
func GetRooms(c *gin.Context) {
	userID, _ := c.Get("user_id") // *Middleware should set user_id
	uid := userID.(uint)

	var rooms []db.Room
	err := db.DB.Where("user_id = ?", uid).Find(&rooms).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

// CreateRoom creates a new room for the authenticated user
func CreateRoom(c *gin.Context) {
	userID, _ := c.Get("user_id") // *Middleware should set user_id
	uid := userID.(uint)

	var req struct {
		Name string `json:"name" binding:"required,min=3,max=50"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// create room
	room := db.Room{
		UserID: uid,
		Name:   req.Name,
	}
	if err := db.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"room": room})
}

// DeleteRoom deletes a room created by the authenticated user
func DeleteRoom(c *gin.Context) {
	userID, _ := c.Get("user_id") // *Middleware should set user_id
	uid := userID.(uint)

	roomID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	// Check if room belong to use
	var room db.Room
	if err := db.DB.Where("id = ? AND user_id = ?", roomID, uid).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// remove 
	if err := db.DB.Delete(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room deleted"})
}
