package handlers

import (
	"net/http"
	"sort"
	"time"

	"GoCall_api/db"

	"github.com/gin-gonic/gin"
)

// ChatHistoryRequest is used to parse query parameters
type ChatHistoryRequest struct {
	WithUser string `form:"with_user" binding:"required"` // UUID of the friend
}

// ChatMessageResponse is used to return messages from DB
type ChatMessageResponse struct {
	ID         uint      `json:"id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Text       string    `json:"text"`
	CreatedAt  time.Time `json:"created_at"`
}

// ConversationResponse represents the latest message preview for one peer.
type ConversationResponse struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Name          string    `json:"name"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
}

// GetChatHistory returns messages between the authenticated user and `with_user`
func GetChatHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var currentUser db.User
	if err := db.DB.First(&currentUser, userID.(uint)).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req ChatHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter 'with_user' is required"})
		return
	}

	var messages []db.Message
	if err := db.DB.
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			currentUser.UserID, req.WithUser, req.WithUser, currentUser.UserID).
		Order("id ASC").
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	var response []ChatMessageResponse
	for _, m := range messages {
		response = append(response, ChatMessageResponse{
			ID:         m.ID,
			SenderID:   m.SenderID,
			ReceiverID: m.ReceiverID,
			Text:       m.Text,
			CreatedAt:  m.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"messages": response})
}

// GetChatConversations returns the latest direct-message preview per peer.
func GetChatConversations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var currentUser db.User
	if err := db.DB.First(&currentUser, userID.(uint)).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var latestMessageIDs []uint
	if err := db.DB.Raw(`
		SELECT MAX(id) AS id
		FROM messages
		WHERE sender_id = ? OR receiver_id = ?
		GROUP BY CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END
	`, currentUser.UserID, currentUser.UserID, currentUser.UserID).Pluck("id", &latestMessageIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	if len(latestMessageIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"conversations": []ConversationResponse{}})
		return
	}

	var messages []db.Message
	if err := db.DB.
		Where("id IN ?", latestMessageIDs).
		Order("created_at DESC, id DESC").
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	conversations := make([]ConversationResponse, 0, len(messages))
	for _, message := range messages {
		peerID := message.SenderID
		if peerID == currentUser.UserID {
			peerID = message.ReceiverID
		}

		var user db.User
		if err := db.DB.Where("user_id = ?", peerID).First(&user).Error; err != nil {
			continue
		}

		conversations = append(conversations, ConversationResponse{
			UserID:        user.UserID,
			Username:      user.Username,
			Name:          user.Name,
			LastMessage:   message.Text,
			LastMessageAt: message.CreatedAt,
		})
	}

	sort.Slice(conversations, func(i, j int) bool {
		return conversations[i].LastMessageAt.After(conversations[j].LastMessageAt)
	})

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}
