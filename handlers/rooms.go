package handlers

import (
	"GoCall_api/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetRooms retrieves the list of rooms created by the authenticated user
func GetRooms(c *gin.Context) {
    userID, ok := c.Get("user_id")
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    userUintID, ok := userID.(uint)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
        return
    }

    var user db.User
    if err := db.DB.First(&user, userUintID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		return
    }

    var rooms []db.Room
    if err := db.DB.Where("user_id = ?", user.UserID).Find(&rooms).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

// CreateRoom creates a new room for the authenticated user
func CreateRoom(c *gin.Context) {
    var req struct {
        Name     string `json:"name" binding:"required,min=3,max=50"`
        Type     string `json:"type" binding:"required,oneof=public private secret"`
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    userID, ok := c.Get("user_id")
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    // Преобразуем userID к uint и получаем UserID как UUID
    userUintID, ok := userID.(uint)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
        return
    }

    var user db.User
    if err := db.DB.First(&user, userUintID).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
        return
    }

    // Create room
    room := db.Room{
        Name:     req.Name,
        UserID:   user.UserID,
        Type:     req.Type,
        Password: req.Password,
    }
    if err := db.DB.Create(&room).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "roomID": room.RoomID,
        "name":   room.Name,
        "type":   room.Type,
    })
}

// AddUserToRoom adds a user to a room
func AddUserToRoom(c *gin.Context) {
	var req struct {
		RoomID string `json:"roomID" binding:"required"`
		UserID string `json:"userID" binding:"required"`
		Role   string `json:"role" binding:"required,oneof=admin member viewer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Add user to room
	member := db.RoomMember{
		RoomID: req.RoomID,
		UserID: req.UserID,
		Role:   req.Role,
	}
	if err := db.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to room"})
}

// GetRoomMembers retrieves the list of members in a room
func GetRoomMembers(c *gin.Context) {
	roomID := c.Param("id")

	var members []db.RoomMember
	if err := db.DB.Where("room_id = ?", roomID).Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

// DeleteRoom deletes a room created by the authenticated user
func DeleteRoom(c *gin.Context) {
    userID, ok := c.Get("user_id")
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    userUintID, ok := userID.(uint)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
        return
    }

    var user db.User
    if err := db.DB.First(&user, userUintID).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
        return
    }

    roomID := c.Param("id")
    var room db.Room
    if err := db.DB.Where("room_id = ? AND user_id = ?", roomID, user.UserID).First(&room).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
        return
    }

    if err := db.DB.Delete(&room).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete room"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Room deleted"})
}

// UpdateRoom updates the room details
func UpdateRoom(c *gin.Context) {
	userID, _ := c.Get("user_id")

	roomID := c.Param("id")
	var req struct {
		Name     string `json:"name" binding:"required,min=3,max=50"`
		Type     string `json:"type" binding:"required,oneof=public private secret"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var room db.Room
	if err := db.DB.Where("room_id = ? AND user_id = ?", roomID, userID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	room.Name = req.Name
	room.Type = req.Type
	room.Password = req.Password

	if err := db.DB.Save(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roomID": room.RoomID, "name": room.Name, "type": room.Type})
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