package handlers

import (
	"net/http"
	"strconv"

	"GoCall_api/db"

	"github.com/gin-gonic/gin"
)

func GetFriends(c *gin.Context) {
	// parse user from JWT

	userID, _ := c.Get("user_id") // *Middleware should set user_id
	uid := userID.(uint)

	var friends []db.User
	err := db.DB.Raw(`
		SELECT u.id, u.username 
		FROM users u
		INNER JOIN friends f ON f.friend_id = u.id
		WHERE f.user_id = ?`, uid).Scan(&friends).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch friends"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"friends": friends})
}

func AddFriend(c *gin.Context) {
	userID, _ := c.Get("user_id") // *Middleware should set user_id
	uid := userID.(uint)

	var req struct {
		FriendUsername string `json:"friend_username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// find friend username
	var friend db.User
	if err := db.DB.Where("username = ?", req.FriendUsername).First(&friend).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend not found"})
		return
	}

	var existingFriend db.Friend
	if err := db.DB.Where("user_id = ? AND friend_id = ?", uid, friend.ID).First(&existingFriend).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Friend already added"})
		return
	}

	// add friend
	newFriend := db.Friend{
		UserID:   uid,
		FriendID: friend.ID,
	}
	if err := db.DB.Create(&newFriend).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add friend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend added"})
}

func RemoveFriend(c *gin.Context) {
	userID, _ := c.Get("user_id") // *Middleware should set user_id
	uid := userID.(uint)

	friendID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend ID"})
		return
	}

	if err := db.DB.Where("user_id = ? AND friend_id = ?", uid, friendID).Delete(&db.Friend{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove friend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend removed"})
}