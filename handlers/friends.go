package handlers

import (
	"net/http"

	"GoCall_api/db"

	"github.com/gin-gonic/gin"
)

func GetFriends(c *gin.Context) {
	// parse user from JWT
	userID := 1 // temp

	rows, err := db.DB.Query("SELECT friend_id FROM friends WHERE user_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch friends"})
		return
	}
	defer rows.Close()

	var friends []int
	for rows.Next() {
		var friendID int
		rows.Scan(&friendID)
		friends = append(friends, friendID)
	}

	c.JSON(http.StatusOK, gin.H{"friends": friends})
}

func AddFriend(c *gin.Context) {
	// adding friend
}

func RemoveFriend(c *gin.Context) {
	// removing friend
}