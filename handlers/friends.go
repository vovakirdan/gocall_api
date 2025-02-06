package handlers

import (
	"net/http"
	"time"

	"GoCall_api/db"

	"github.com/gin-gonic/gin"
)

// Struct for getting friend data with online status
type FriendUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	IsOnline bool   `json:"is_online"`
	UserID string `json:"user_id"`
	IsPinned  bool   `json:"is_pinned"`
	CreatedAt time.Time `json:"created_at"`
}

// GetFriends returns all accepted friends
func GetFriends(c *gin.Context) {
	userID, exists := c.Get("user_id")
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

	var friends []FriendUser
	err := db.DB.Raw(`
		SELECT DISTINCT u.id, u.username, u.is_online, u.user_id, f.is_pinned, f.created_at
		FROM users u
		JOIN friends f
		ON (f.friend_id = u.user_id AND f.user_id = ?)
	`, currentUser.UserID, currentUser.UserID).Scan(&friends).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch friends"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"friends": friends})
}

// AddFriend (direct add; bypasses friend request flow)
func AddFriend(c *gin.Context) {
	userID, exists := c.Get("user_id")
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

	var existingFriend db.Friend
	if err := db.DB.Where("user_id = ? AND friend_id = ?", currentUser.UserID, friend.UserID).First(&existingFriend).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already friends"})
		return
	}

	newFriend1 := db.Friend{
		UserID:   currentUser.UserID,
		FriendID: friend.UserID,
	}
	if err := db.DB.Create(&newFriend1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add friend"})
		return
	}

	newFriend2 := db.Friend{
		UserID:   friend.UserID,
		FriendID: currentUser.UserID,
	}
	if err := db.DB.Create(&newFriend2).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add friend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend added", "friend": friend.Username})
}

// RemoveFriend
func RemoveFriend(c *gin.Context) {
	userID, exists := c.Get("user_id")
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove friend"})
		return
	}
	if err := db.DB.Where("user_id = ? AND friend_id = ?", friend.UserID, currentUser.UserID).Delete(&db.Friend{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove friend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend removed"})
}

// Send friend request
func RequestFriend(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid := userID.(uint)

	var currentUser db.User
	if err := db.DB.First(&currentUser, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find user"})
		return
	}

	var req struct {
		ToUsername string `json:"to_username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var toUser db.User
	if err := db.DB.Where("username = ?", req.ToUsername).First(&toUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target user not found"})
		return
	}

	// Check if a request already exists
	var fr db.FriendRequest
	if err := db.DB.Where("from_user_id = ? AND to_user_id = ? AND status = 'pending'",
		currentUser.UserID, toUser.UserID).First(&fr).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Friend request already sent"})
		return
	}

	newRequest := db.FriendRequest{
		FromUserID: currentUser.UserID,
		ToUserID:   toUser.UserID,
		Status:     "pending",
	}
	if err := db.DB.Create(&newRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create friend request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request sent"})
}

// Accept friend request
func AcceptFriendRequest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid := userID.(uint)

	var currentUser db.User
	if err := db.DB.First(&currentUser, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find user"})
		return
	}

	var req struct {
		RequestID uint `json:"request_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var fr db.FriendRequest
	if err := db.DB.First(&fr, req.RequestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found"})
		return
	}

	if fr.ToUserID != currentUser.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to accept this request"})
		return
	}
	if fr.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"error": "Request is not pending"})
		return
	}

	fr.Status = "accepted"
	if err := db.DB.Save(&fr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not accept request"})
		return
	}

	// Create friendship
	newFriend1 := db.Friend{
		UserID:   fr.FromUserID,
		FriendID: fr.ToUserID,
		CreatedAt: time.Now(),
	}
	db.DB.Create(&newFriend1)
	newFriend2 := db.Friend{
		UserID:   fr.ToUserID,
		FriendID: fr.FromUserID,
		CreatedAt: time.Now(),
	}
	db.DB.Create(&newFriend2)

	c.JSON(http.StatusOK, gin.H{"message": "Friend request accepted"})
}

// Decline friend request
func DeclineFriendRequest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid := userID.(uint)

	var currentUser db.User
	if err := db.DB.First(&currentUser, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find user"})
		return
	}

	var req struct {
		RequestID uint `json:"request_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var fr db.FriendRequest
	if err := db.DB.First(&fr, req.RequestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found"})
		return
	}

	if fr.ToUserID != currentUser.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed to decline this request"})
		return
	}
	if fr.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"error": "Request is not pending"})
		return
	}

	fr.Status = "declined"
	if err := db.DB.Save(&fr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not decline request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request declined"})
}

// Get friend requests
func GetFriendRequests(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid := userID.(uint)

	var currentUser db.User
	if err := db.DB.First(&currentUser, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find user"})
		return
	}

	var friendRequests []db.FriendRequest
	if err := db.DB.Where("to_user_id = ? AND status = 'pending'", currentUser.UserID).Find(&friendRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch friend requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"friend_requests": friendRequests})
}

// PinFriend sets is_pinned = true on the existing friendship row
// request body: { "friend_id": 123 } - this is the DB's numeric ID, not the UUID
func PinFriend(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid := userID.(uint)

	// We find the user by its DB ID
	var currentUser db.User
	if err := db.DB.First(&currentUser, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find user"})
		return
	}

	// Read friend_id from the JSON body
	var req struct {
		FriendID uint `json:"friend_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Now let's find that friend user by DB ID
	var friendUser db.User
	if err := db.DB.Where("id = ?", req.FriendID).First(&friendUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend user not found"})
		return
	}

	// They must already be in the Friend table or else not friends.
	var friendship db.Friend
	err := db.DB.Where("user_id = ? AND friend_id = ?", currentUser.UserID, friendUser.UserID).First(&friendship).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "You are not friends with this user"})
		return
	}

	// If already pinned, do nothing or return conflict
	if friendship.IsPinned {
		c.JSON(http.StatusConflict, gin.H{"error": "Friend is already pinned"})
		return
	}

	// Otherwise, update IsPinned
	friendship.IsPinned = true
	if err := db.DB.Save(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not pin friend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend pinned"})
}

// UnpinFriend sets is_pinned = false on the existing friendship row
// request body: { "friend_id": 123 }
func UnpinFriend(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid := userID.(uint)

	var currentUser db.User
	if err := db.DB.First(&currentUser, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find user"})
		return
	}

	var req struct {
		FriendID uint `json:"friend_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var friendUser db.User
	if err := db.DB.Where("id = ?", req.FriendID).First(&friendUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend user not found"})
		return
	}

	var friendship db.Friend
	err := db.DB.Where("user_id = ? AND friend_id = ?", currentUser.UserID, friendUser.UserID).First(&friendship).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend is not pinned or not a friend"})
		return
	}

	// If not pinned, return conflict or do nothing
	if !friendship.IsPinned {
		c.JSON(http.StatusConflict, gin.H{"error": "Friend is not pinned"})
		return
	}

	friendship.IsPinned = false
	if err := db.DB.Save(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not unpin friend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend unpinned"})
}

// GetPinnedFriends returns a list of pinned friendships
func GetPinnedFriends(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid := userID.(uint)

	var currentUser db.User
	if err := db.DB.First(&currentUser, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find user"})
		return
	}

	// Get all pinned friends for the current user
	var pinnedFriends []db.Friend
	if err := db.DB.Where("user_id = ? AND is_pinned = ?", currentUser.UserID, true).Find(&pinnedFriends).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch pinned friends"})
		return
	}

	var result []FriendUser
	for _, pf := range pinnedFriends {
		var friendUser db.User
		if err := db.DB.Where("user_id = ?", pf.FriendID).First(&friendUser).Error; err == nil {
			result = append(result, FriendUser{
				ID:        pf.ID,
				UserID:    pf.FriendID,
				Username:  friendUser.Username,
				IsPinned:  pf.IsPinned,
				CreatedAt: pf.CreatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"pinned_friends": result})
}

