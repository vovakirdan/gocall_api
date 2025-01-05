package handlers

import (
	"net/http"

	"GoCall_api/db"

	"github.com/gin-gonic/gin"
)

func GetFriends(c *gin.Context) {
    // Parse user ID from JWT middleware
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    uid := userID.(uint)

    // Find the current user
    var currentUser db.User
    if err := db.DB.First(&currentUser, uid).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch authenticated user"})
        return
    }

    // Fetch friends where the user is either the initiator or the friend
    var friends []db.User
    err := db.DB.Raw(`
        SELECT DISTINCT u.id, u.username 
        FROM users u
        INNER JOIN friends f 
        ON (f.friend_id = u.user_id AND f.user_id = ?) 
        OR (f.user_id = u.user_id AND f.friend_id = ?)
    `, currentUser.UserID, currentUser.UserID).Scan(&friends).Error

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch friends"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"friends": friends})
}

func AddFriend(c *gin.Context) {
    userID, exists := c.Get("user_id") // Middleware sets user_id
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    uid := userID.(uint)

    // Find the authenticated user's UUID
    var currentUser db.User
    if err := db.DB.First(&currentUser, uid).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch authenticated user"})
        return
    }

    var req struct {
        FriendUsername string `json:"friend_username" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input. 'friend_username' is required"})
        return
    }

    var friend db.User
    if err := db.DB.Where("username = ?", req.FriendUsername).First(&friend).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User with this username not found"})
        return
    }

    var existingFriend db.Friend
    if err := db.DB.Where("user_id = ? AND friend_id = ?", currentUser.UserID, friend.UserID).First(&existingFriend).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "This user is already your friend"})
        return
    }

    newFriendForUser1 := db.Friend{
        UserID:   currentUser.UserID,
        FriendID: friend.UserID,
    }
    if err := db.DB.Create(&newFriendForUser1).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add friend for user 1"})
        return
    }

    newFriendForUser2 := db.Friend{
        UserID:   friend.UserID,
        FriendID: currentUser.UserID,
    }
    if err := db.DB.Create(&newFriendForUser2).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add friend for user 2"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Friend successfully added", "friend": friend.Username})
}

func RemoveFriend(c *gin.Context) {
    userID, exists := c.Get("user_id") // Middleware sets user_id
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    uid := userID.(uint)

    var currentUser db.User
    if err := db.DB.First(&currentUser, uid).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch authenticated user"})
        return
    }

    var req struct {
        FriendUsername string `json:"friend_username" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input. 'friend_username' is required"})
        return
    }

	var friend db.User
    if err := db.DB.Where("username = ?", req.FriendUsername).First(&friend).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User with this username not found"})
        return
    }

    if err := db.DB.Where("user_id = ? AND friend_id = ?", currentUser.UserID, friend.UserID).Delete(&db.Friend{}).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove friend for user 1"})
        return
    }

    if err := db.DB.Where("user_id = ? AND friend_id = ?", friend.UserID, currentUser.UserID).Delete(&db.Friend{}).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove friend for user 2"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Friend removed"})
}