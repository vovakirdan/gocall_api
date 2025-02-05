package handlers

import (
	"net/http"

	"GoCall_api/db"

	"github.com/gin-gonic/gin"
)

var users []struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Name     string `json:"name"`
}

// GetUserID returns the authenticated user's UUID
func GetUserID(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var user db.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"userID": user.UserID})
}

func SearchUsers(c *gin.Context) {
    query := c.Query("q")
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
        return
    }

    if err := db.DB.Model(&db.User{}).Select("ID, Username, Name").Where("username LIKE ?", "%"+query+"%").Limit(10).Find(&users).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUserByUUID returns a user by their UUID
func GetUserByUUID(c *gin.Context) {
	userID := c.Param("uuid")
	if err := db.DB.Model(&db.User{}).Where("user_id = ?", userID).First(&users).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": users})
}

// GetUserByToken returns the user info by token
func GetUserByToken(c *gin.Context) {
    userID := c.MustGet("user_id").(uint) // todo rename user_id to id; userID is the user's UUID, id is the user's ID (integer)
    var user db.User

    if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":        user.ID,
        "user_id":   user.UserID,
        "username":  user.Username,
        "name":      user.Name,
        "email":     user.Email,
        "is_online": user.IsOnline,
        "created_at": user.CreatedAt,
    })
}
