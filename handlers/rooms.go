package handlers

import (
	"GoCall_api/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetRooms retrieves the list of rooms created by the authenticated user
func GetRooms(c *gin.Context) {
	userID, _ := c.Get("user_id") // *Middleware should set user_id
	var user db.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		return
	}

	var rooms []db.Room
	err := db.DB.Where("user_id = ?", user.UserID).Find(&rooms).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

// CreateRoom creates a new room for the authenticated user
func CreateRoom(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required,min=3,max=50"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	// get userID (uuid) from users
	var user db.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		return
	}
	// create room
	room := db.Room{
		Name:   req.Name,
		UserID: user.UserID,
	}
	if err := db.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roomID": room.RoomID,
		"name": room.Name,
		"userID": room.UserID,
	})
}

// DeleteRoom deletes a room created by the authenticated user
func DeleteRoom(c *gin.Context) {
	userID, _ := c.Get("user_id") // *Middleware should set user_id

	roomID := c.Param("id")

	var user db.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		return
	}

	// Check if room belong to use
	var room db.Room
	if err := db.DB.Where("room_id = ? AND user_id = ?", roomID, user.UserID).First(&room).Error; err != nil {
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

// UpdateRoom updates the room details
func UpdateRoom(c *gin.Context) {
    userID, _ := c.Get("user_id")
    uid := userID.(uint)

    roomID := c.Param("id")
    var req struct {
        Name string `json:"name" binding:"required,min=3,max=50"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    var room db.Room
    if err := db.DB.Where("room_id = ? AND user_id = ?", roomID, uid).First(&room).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
        return
    }

    room.Name = req.Name
    if err := db.DB.Save(&room).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"roomID": room.RoomID, "name": room.Name})
}

// GetRoomByName retrieves room details by name
func GetRoomByName(c *gin.Context) {
	var room db.Room
	roomName := c.Query("name")
	if err := db.DB.Where("name = ?", roomName).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"roomID": room.RoomID, "name": room.Name})
}