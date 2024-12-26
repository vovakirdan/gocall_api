package main

import (
	"os"
	"errors"
	"io/fs"
	"log"

	"GoCall_api/db"
	"GoCall_api/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// check path data exists
	dataExists, err := exists("./data")
	if err != nil {
		log.Fatal(err)
	}
	if !dataExists {
		_ = os.Mkdir("./data", 0700)
	}
	db.InitDatabase("./data/gocall.db")

	router := gin.Default()

	api := router.Group("/api")
	{
		api.POST("/auth/login", handlers.Login)
		api.POST("/auth/register", handlers.Register)
		api.GET("/friends", handlers.GetFriends)
		api.POST("/friends/add", handlers.AddFriend)
		api.DELETE("/friends/remove", handlers.RemoveFriend)
		api.GET("/rooms", handlers.GetRooms)
		api.POST("/rooms/create", handlers.CreateRoom)
		api.DELETE("/rooms/:id", handlers.DeleteRoom)
	}

	router.Run(":8080")
}

func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil {
        return true, nil
    }
    if errors.Is(err, fs.ErrNotExist) {
        return false, nil
    }
    return false, err
}