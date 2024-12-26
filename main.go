package main

import (
	"os"
	"errors"
	"io/fs"
	"log"

	"GoCall_api/db"
	"GoCall_api/handlers"
	"GoCall_api/utils"

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
		// with no auth
		api.POST("/auth/login", handlers.Login)
		api.POST("/auth/register", handlers.Register)

		// With auth
		protected := api.Group("/")
		protected.Use(utils.JWTMiddleware())
		{
			protected.GET("/friends", handlers.GetFriends)
			protected.POST("/friends/add", handlers.AddFriend)
			protected.DELETE("/friends/remove", handlers.RemoveFriend)
			protected.GET("/rooms", handlers.GetRooms)
			protected.POST("/rooms/create", handlers.CreateRoom)
			protected.DELETE("/rooms/:id", handlers.DeleteRoom)
		}
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