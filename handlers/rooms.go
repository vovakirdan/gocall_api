package handlers

import (
	"GoCall_api/db"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RoomExists reports whether a room with the given ID exists.
func RoomExists(c *gin.Context) {
	roomID := c.Param("id")
	var room db.Room
	if err := db.DB.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"exists": false, "error": "Room not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"exists": true})
}

// GetAllPublicRooms returns all public rooms without requiring auth.
func GetAllPublicRooms(c *gin.Context) {
	var rooms []db.Room
	err := db.DB.Where("type = ?", "public").Find(&rooms).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public rooms"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

// Get rooms where the authenticated user is a member.
func GetMyRooms(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	uid := userID.(uint)

	var currentUser db.User
	if err := db.DB.First(&currentUser, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	var rooms []db.Room
	if err := db.DB.
		Table("rooms").
		Joins("JOIN room_members ON room_members.room_id = rooms.room_id").
		Where("room_members.user_id = ?", currentUser.UserID).
		Order("rooms.created_at DESC").
		Find(&rooms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

// CreateRoom creates a new room and adds the creator as a member.
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	uid := userID.(uint)

	var user db.User
	if err := db.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	room := db.Room{
		UserID:   user.UserID,
		Name:     req.Name,
		Type:     req.Type,
		Password: req.Password,
	}
	if err := db.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
		return
	}

	member := db.RoomMember{
		RoomID: room.RoomID,
		UserID: user.UserID,
		Role:   "creator",
	}
	db.DB.Create(&member)

	c.JSON(http.StatusOK, gin.H{"roomID": room.RoomID, "name": room.Name, "type": room.Type})
}

// GetOrCreateDirectRoom returns the existing direct room between two friends or creates one.
func GetOrCreateDirectRoom(c *gin.Context) {
	var req struct {
		FriendUserID string `json:"friend_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var currentUser db.User
	if err := db.DB.First(&currentUser, userID.(uint)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	if req.FriendUserID == currentUser.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create a direct room with yourself"})
		return
	}

	var friend db.User
	if err := db.DB.Where("user_id = ?", req.FriendUserID).First(&friend).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend user not found"})
		return
	}

	if !areFriends(currentUser.UserID, friend.UserID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Users must be friends"})
		return
	}

	memberIDs := []string{currentUser.UserID, friend.UserID}
	sort.Strings(memberIDs)
	roomName := "__direct__:" + strings.Join(memberIDs, ":")

	var room db.Room
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		room = db.Room{
			UserID: currentUser.UserID,
			Name:   roomName,
			Type:   "secret",
		}

		return tx.Where("name = ? AND type = ?", roomName, "secret").
			Attrs(db.Room{
				UserID: currentUser.UserID,
				Name:   roomName,
				Type:   "secret",
			}).
			FirstOrCreate(&room).Error
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch or create direct room"})
		return
	}

	for index, memberID := range memberIDs {
		role := "member"
		if memberID == room.UserID || index == 0 {
			role = "creator"
		}

		member := db.RoomMember{}
		memberErr := db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, memberID).First(&member).Error
		if memberErr == nil {
			continue
		}
		if memberErr != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room member"})
			return
		}

		if err := db.DB.Create(&db.RoomMember{
			RoomID: room.RoomID,
			UserID: memberID,
			Role:   role,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add direct room member"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"room": room})
}

// GetRoomByID returns room details, enforcing visibility by room type.
func GetRoomByID(c *gin.Context) {
	roomID := c.Param("id")

	var room db.Room
	if err := db.DB.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	if room.Type == "public" {
		c.JSON(http.StatusOK, gin.H{"room": room})
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not allowed"})
		return
	}
	uid := userID.(uint)

	var currentUser db.User
	if err := db.DB.First(&currentUser, uid).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var member db.RoomMember
	err := db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, currentUser.UserID).
		First(&member).Error
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Room is private or secret"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"room": room})
}

// UpdateRoom updates mutable room fields for creator or admins.
func UpdateRoom(c *gin.Context) {
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

	userID, _ := c.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	uid := userID.(uint)

	var currentUser db.User
	db.DB.First(&currentUser, uid)

	var room db.Room
	if err := db.DB.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	var member db.RoomMember
	if db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, currentUser.UserID).
		First(&member).Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member"})
		return
	}
	if member.Role != "creator" && member.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "No permissions"})
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

// DeleteRoom removes a room and its related membership/invite records.
func DeleteRoom(c *gin.Context) {
	roomID := c.Param("id")

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	uid := userID.(uint)

	var currentUser db.User
	db.DB.First(&currentUser, uid)

	var room db.Room
	if err := db.DB.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	var member db.RoomMember
	if db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, currentUser.UserID).First(&member).Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member"})
		return
	}
	if member.Role != "creator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only creator can delete"})
		return
	}

	db.DB.Where("room_id = ?", room.RoomID).Delete(&db.RoomMember{})
	db.DB.Where("room_id = ?", room.RoomID).Delete(&db.RoomInvite{})
	db.DB.Delete(&room)

	c.JSON(http.StatusOK, gin.H{"message": "Room deleted"})
}

// MakeRoomAdmin promotes an existing room member to admin.
func MakeRoomAdmin(c *gin.Context) {
	var req struct {
		UserToAdmin string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	roomID := c.Param("id")

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	uid := userID.(uint)

	var currentUser db.User
	db.DB.First(&currentUser, uid)

	var room db.Room
	if err := db.DB.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	var member db.RoomMember
	if db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, currentUser.UserID).
		First(&member).Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member"})
		return
	}
	if member.Role != "creator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only creator can assign admin"})
		return
	}

	var targetMember db.RoomMember
	if db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, req.UserToAdmin).
		First(&targetMember).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target user not in room"})
		return
	}

	targetMember.Role = "admin"
	db.DB.Save(&targetMember)

	c.JSON(http.StatusOK, gin.H{"message": "User assigned as admin"})
}

// InviteUserToRoom creates a pending invitation for a registered user.
func InviteUserToRoom(c *gin.Context) {
	var req struct {
		RoomID   string `json:"roomID" binding:"required"`
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	uid := userID.(uint)

	var inviter db.User
	if err := db.DB.First(&inviter, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Inviter not found"})
		return
	}

	var room db.Room
	if err := db.DB.Where("room_id = ?", req.RoomID).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	var member db.RoomMember
	if db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, inviter.UserID).
		First(&member).Error != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member"})
		return
	}
	if member.Role != "creator" && member.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "No permission to invite"})
		return
	}

	var invitedUser db.User
	if err := db.DB.Where("username = ?", req.Username).First(&invitedUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var existingInvite db.RoomInvite
	if err := db.DB.Where("room_id = ? AND invited_user_id = ? AND status = 'pending'",
		room.RoomID, invitedUser.UserID).First(&existingInvite).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Invite already pending"})
		return
	}

	invite := db.RoomInvite{
		RoomID:        room.RoomID,
		InviterUserID: inviter.UserID,
		InvitedUserID: invitedUser.UserID,
		Status:        "pending",
	}
	db.DB.Create(&invite)

	c.JSON(http.StatusOK, gin.H{"message": "Invitation sent"})
}

// AcceptRoomInvite accepts a pending invitation and joins the room.
func AcceptRoomInvite(c *gin.Context) {
	var req struct {
		InviteID uint `json:"invite_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	uid := userID.(uint)

	var user db.User
	db.DB.First(&user, uid)

	var invite db.RoomInvite
	if err := db.DB.First(&invite, req.InviteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invite not found"})
		return
	}

	if invite.InvitedUserID != user.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your invite"})
		return
	}
	if invite.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"error": "Invite not pending"})
		return
	}

	invite.Status = "accepted"
	db.DB.Save(&invite)

	var existingMember db.RoomMember
	if err := db.DB.Where("room_id = ? AND user_id = ?", invite.RoomID, user.UserID).
		First(&existingMember).Error; err != nil {
		member := db.RoomMember{
			RoomID: invite.RoomID,
			UserID: user.UserID,
			Role:   "member",
		}
		db.DB.Create(&member)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invite accepted"})
}

// DeclineRoomInvite declines a pending room invitation.
func DeclineRoomInvite(c *gin.Context) {
	var req struct {
		InviteID uint `json:"invite_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	uid := userID.(uint)

	var user db.User
	db.DB.First(&user, uid)

	var invite db.RoomInvite
	if err := db.DB.First(&invite, req.InviteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invite not found"})
		return
	}

	if invite.InvitedUserID != user.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your invite"})
		return
	}
	if invite.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"error": "Invite not pending"})
		return
	}

	invite.Status = "declined"
	db.DB.Save(&invite)

	c.JSON(http.StatusOK, gin.H{"message": "Invite declined"})
}

// GetRoomInvites returns pending and accepted invites for the authenticated user.
func GetRoomInvites(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	uid := userID.(uint)

	var user db.User
	db.DB.First(&user, uid)

	var invites []db.RoomInvite
	err := db.DB.Where("invited_user_id = ? AND (status = 'pending' OR status = 'accepted')",
		user.UserID).Find(&invites).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invites": invites})
}
