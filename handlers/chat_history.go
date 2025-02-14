package handlers

import (
    "net/http"
    "time"

    "GoCall_api/db"
    "GoCall_api/utils"

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

// GetChatHistory returns messages between the authenticated user and `with_user`
func GetChatHistory(c *gin.Context) {
    tokenString := c.Query("token")
    if tokenString == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
        return
    }
    dbUserID, err := utils.DecodeJWT(tokenString)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
        return
    }

    // Найдём пользователя (чтобы получить его UUID)
    var currentUser db.User
    if err := db.DB.First(&currentUser, dbUserID).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
        return
    }

    // 2. Считываем параметр `with_user`
    var req ChatHistoryRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter 'with_user' is required"})
        return
    }

    // 3. Выбираем сообщения, где (sender = currentUser.UserID, receiver = req.WithUser)
    //    или (sender = req.WithUser, receiver = currentUser.UserID)
    var messages []db.Message
    if err := db.DB.
        Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
            currentUser.UserID, req.WithUser, req.WithUser, currentUser.UserID).
        Order("id ASC").
        Find(&messages).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
        return
    }

    // 4. Преобразуем в удобную структуру
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

    // 5. Отдаём результат
    c.JSON(http.StatusOK, gin.H{"messages": response})
}