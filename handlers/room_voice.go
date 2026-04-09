package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"GoCall_api/db"

	"github.com/gin-gonic/gin"
	"github.com/livekit/protocol/auth"
)

type roomMemberState struct {
	ID       uint   `json:"id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	IsOnline bool   `json:"is_online"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

type roomVoiceParticipantState struct {
	ID              uint   `json:"id"`
	UserID          string `json:"user_id"`
	Username        string `json:"username"`
	Name            string `json:"name"`
	IsOnline        bool   `json:"is_online"`
	IsMicEnabled    bool   `json:"is_mic_enabled"`
	IsCameraEnabled bool   `json:"is_camera_enabled"`
	IsScreenSharing bool   `json:"is_screen_sharing"`
	JoinedAt        string `json:"joined_at"`
	UpdatedAt       string `json:"updated_at"`
}

type roomStateRoom struct {
	ID        uint   `json:"id"`
	RoomID    string `json:"room_id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}

type roomStateResponse struct {
	Room              roomStateRoom               `json:"room"`
	Members           []roomMemberState           `json:"members"`
	VoiceParticipants []roomVoiceParticipantState `json:"voice_participants"`
	InVoice           bool                        `json:"in_voice"`
}

type roomVoiceCredentialsResponse struct {
	URL      string `json:"url"`
	Token    string `json:"token"`
	RoomName string `json:"room_name"`
	Identity string `json:"identity"`
	Name     string `json:"name"`
}

func resolveRoomByParam(idOrRoomID string) (*db.Room, error) {
	var room db.Room
	if err := db.DB.Where("room_id = ?", idOrRoomID).First(&room).Error; err == nil {
		return &room, nil
	}

	numericID, err := strconv.ParseUint(idOrRoomID, 10, 64)
	if err != nil {
		return nil, db.DB.Where("room_id = ?", idOrRoomID).First(&room).Error
	}

	if err := db.DB.First(&room, uint(numericID)).Error; err != nil {
		return nil, err
	}

	return &room, nil
}

func getAuthenticatedDBUser(c *gin.Context) (*db.User, bool) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return nil, false
	}

	var currentUser db.User
	if err := db.DB.First(&currentUser, userID.(uint)).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return nil, false
	}

	return &currentUser, true
}

func resolveLiveKitPublicURL(c *gin.Context, fallback string) string {
	if fallback != "" {
		return fallback
	}

	scheme := "ws"
	if strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") || c.Request.TLS != nil {
		scheme = "wss"
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	if host == "" {
		return ""
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

func ensureRoomMember(room *db.Room, user *db.User) (*db.RoomMember, error) {
	var member db.RoomMember
	err := db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, user.UserID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// JoinRoom ensures the authenticated user is a member of the room.
// JoinRoom ensures the authenticated user is a member of the room.
// Public rooms auto-create membership, private/secret rooms require existing membership.
func JoinRoom(c *gin.Context) {
	currentUser, ok := getAuthenticatedDBUser(c)
	if !ok {
		return
	}

	room, err := resolveRoomByParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	var member db.RoomMember
	if err := db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, currentUser.UserID).First(&member).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Already in room", "room_id": room.RoomID})
		return
	}

	if room.Type != "public" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Room membership is required"})
		return
	}

	member = db.RoomMember{
		RoomID: room.RoomID,
		UserID: currentUser.UserID,
		Role:   "member",
	}
	if err := db.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joined room", "room_id": room.RoomID})
}

// GetRoomState returns room metadata, members, and current voice participants.
func GetRoomState(c *gin.Context) {
	currentUser, ok := getAuthenticatedDBUser(c)
	if !ok {
		return
	}

	room, err := resolveRoomByParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	if _, err := ensureRoomMember(room, currentUser); err != nil {
		if room.Type != "public" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not a room member"})
			return
		}
	}

	var members []db.RoomMember
	if err := db.DB.Where("room_id = ?", room.RoomID).Order("joined_at ASC").Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room members"})
		return
	}

	memberStates := make([]roomMemberState, 0, len(members))
	for _, member := range members {
		var user db.User
		if err := db.DB.Where("user_id = ?", member.UserID).First(&user).Error; err != nil {
			continue
		}

		memberStates = append(memberStates, roomMemberState{
			ID:       user.ID,
			UserID:   user.UserID,
			Username: user.Username,
			Name:     user.Name,
			IsOnline: user.IsOnline,
			Role:     member.Role,
			JoinedAt: member.JoinedAt.Format(http.TimeFormat),
		})
	}

	var voiceParticipants []db.RoomVoiceParticipant
	if err := db.DB.Where("room_id = ?", room.RoomID).Order("joined_at ASC").Find(&voiceParticipants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room voice state"})
		return
	}

	voiceStates := make([]roomVoiceParticipantState, 0, len(voiceParticipants))
	inVoice := false
	for _, voiceParticipant := range voiceParticipants {
		var user db.User
		if err := db.DB.Where("user_id = ?", voiceParticipant.UserID).First(&user).Error; err != nil {
			continue
		}

		if voiceParticipant.UserID == currentUser.UserID {
			inVoice = true
		}

		voiceStates = append(voiceStates, roomVoiceParticipantState{
			ID:              user.ID,
			UserID:          user.UserID,
			Username:        user.Username,
			Name:            user.Name,
			IsOnline:        user.IsOnline,
			IsMicEnabled:    voiceParticipant.IsMicEnabled,
			IsCameraEnabled: voiceParticipant.IsCameraEnabled,
			IsScreenSharing: voiceParticipant.IsScreenSharing,
			JoinedAt:        voiceParticipant.JoinedAt.Format(http.TimeFormat),
			UpdatedAt:       voiceParticipant.UpdatedAt.Format(http.TimeFormat),
		})
	}

	c.JSON(http.StatusOK, roomStateResponse{
		Room: roomStateRoom{
			ID:        room.ID,
			RoomID:    room.RoomID,
			UserID:    room.UserID,
			Name:      room.Name,
			Type:      room.Type,
			CreatedAt: room.CreatedAt.Format(http.TimeFormat),
		},
		Members:           memberStates,
		VoiceParticipants: voiceStates,
		InVoice:           inVoice,
	})
}

// JoinRoomVoice adds the authenticated user to room-scoped voice presence with media disabled by default.
func JoinRoomVoice(c *gin.Context) {
	currentUser, ok := getAuthenticatedDBUser(c)
	if !ok {
		return
	}

	room, err := resolveRoomByParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	if _, err := ensureRoomMember(room, currentUser); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a room member"})
		return
	}

	var existing db.RoomVoiceParticipant
	if err := db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, currentUser.UserID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":           "Already in room voice",
			"room_id":           room.RoomID,
			"user_id":           currentUser.UserID,
			"is_mic_enabled":    existing.IsMicEnabled,
			"is_camera_enabled": existing.IsCameraEnabled,
			"is_screen_sharing": existing.IsScreenSharing,
		})
		return
	}

	voiceParticipant := db.RoomVoiceParticipant{
		RoomID:          room.RoomID,
		UserID:          currentUser.UserID,
		IsMicEnabled:    false,
		IsCameraEnabled: false,
		IsScreenSharing: false,
	}
	if err := db.DB.Create(&voiceParticipant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join room voice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Joined room voice",
		"room_id":           room.RoomID,
		"user_id":           currentUser.UserID,
		"is_mic_enabled":    false,
		"is_camera_enabled": false,
		"is_screen_sharing": false,
	})
}

// LeaveRoomVoice removes the authenticated user from room-scoped voice presence.
func LeaveRoomVoice(c *gin.Context) {
	currentUser, ok := getAuthenticatedDBUser(c)
	if !ok {
		return
	}

	room, err := resolveRoomByParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	result := db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, currentUser.UserID).Delete(&db.RoomVoiceParticipant{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave room voice"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User is not in room voice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Left room voice", "room_id": room.RoomID})
}

// UpdateRoomVoiceMedia updates the authenticated user's media flags while they are in room voice.
func UpdateRoomVoiceMedia(c *gin.Context) {
	currentUser, ok := getAuthenticatedDBUser(c)
	if !ok {
		return
	}

	room, err := resolveRoomByParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	var req struct {
		IsMicEnabled    *bool `json:"is_mic_enabled"`
		IsCameraEnabled *bool `json:"is_camera_enabled"`
		IsScreenSharing *bool `json:"is_screen_sharing"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var voiceParticipant db.RoomVoiceParticipant
	if err := db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, currentUser.UserID).First(&voiceParticipant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User is not in room voice"})
		return
	}

	if req.IsMicEnabled != nil {
		voiceParticipant.IsMicEnabled = *req.IsMicEnabled
	}
	if req.IsCameraEnabled != nil {
		voiceParticipant.IsCameraEnabled = *req.IsCameraEnabled
	}
	if req.IsScreenSharing != nil {
		voiceParticipant.IsScreenSharing = *req.IsScreenSharing
	}

	if err := db.DB.Save(&voiceParticipant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room voice media state"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Updated room voice media state",
		"room_id":           room.RoomID,
		"user_id":           currentUser.UserID,
		"is_mic_enabled":    voiceParticipant.IsMicEnabled,
		"is_camera_enabled": voiceParticipant.IsCameraEnabled,
		"is_screen_sharing": voiceParticipant.IsScreenSharing,
	})
}

// GetRoomVoiceCredentials returns LiveKit credentials for a room-scoped voice participant.
func GetRoomVoiceCredentials(c *gin.Context) {
	currentUser, ok := getAuthenticatedDBUser(c)
	if !ok {
		return
	}

	room, err := resolveRoomByParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	var voiceParticipant db.RoomVoiceParticipant
	if err := db.DB.Where("room_id = ? AND user_id = ?", room.RoomID, currentUser.UserID).First(&voiceParticipant).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Join room voice before requesting credentials"})
		return
	}

	livekitURL := resolveLiveKitPublicURL(c, os.Getenv("LIVEKIT_URL"))
	livekitAPIKey := os.Getenv("LIVEKIT_API_KEY")
	livekitAPISecret := os.Getenv("LIVEKIT_API_SECRET")
	if livekitURL == "" || livekitAPIKey == "" || livekitAPISecret == "" {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "LiveKit is not configured. Set LIVEKIT_URL, LIVEKIT_API_KEY, and LIVEKIT_API_SECRET.",
		})
		return
	}

	canPublish := true
	canSubscribe := true
	canPublishData := true
	token := auth.NewAccessToken(livekitAPIKey, livekitAPISecret).
		SetIdentity(currentUser.UserID).
		SetName(currentUser.Username).
		AddGrant(&auth.VideoGrant{
			RoomJoin:       true,
			Room:           room.RoomID,
			CanPublish:     &canPublish,
			CanSubscribe:   &canSubscribe,
			CanPublishData: &canPublishData,
		})

	jwt, err := token.ToJWT()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate LiveKit token"})
		return
	}

	c.JSON(http.StatusOK, roomVoiceCredentialsResponse{
		URL:      livekitURL,
		Token:    jwt,
		RoomName: room.RoomID,
		Identity: currentUser.UserID,
		Name:     currentUser.Username,
	})
}
