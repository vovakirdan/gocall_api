package main

import (
	"os"
	"errors"
	"io/fs"
	"log"
	"strings"

	"GoCall_api/db"
	"GoCall_api/handlers"
	"GoCall_api/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// ENV INIT
	err := utils.CheckEnvLoaded(); if err != nil {
		log.Fatal(err)
	}
	// --------------------------------
	// DATABASE INIT
	// check path data exists
	dataExists, err := exists("./data")
	if err != nil {
		log.Fatal(err)
	}
	if !dataExists {
		_ = os.Mkdir("./data", 0700)
	}
	db.InitDatabase("./data/gocall.db")
	// --------------------------------
	// VALIDATOR INIT
	handlers.InitValidator()
	// --------------------------------
	router := gin.Default()

	router.Use(cors.New(cors.Config{
        AllowOrigins:     strings.Split(os.Getenv("ALLOW_ORIGINS"), ","), // Allow sources. By default: http://127.0.0.1:1420 http://localhost:1420
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:	  []string{"Origin", "Content-Type", "Authorization", "X-Client-Type"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    }))

	api := router.Group("/api")
	{
		// with no auth
		api.POST("/auth/login", handlers.Login)
		api.POST("/auth/register", handlers.Register)
		api.POST("/auth/refresh", utils.RefreshToken)
		api.GET("/rooms/:id/exists", handlers.RoomExists)  //! temp remove

		// With auth
		protected := api.Group("/")
		protected.Use(utils.JWTMiddleware())
		{
			// ----------USERS--------------------------
			protected.GET("/user/id", handlers.GetUserID)
			protected.GET("/friends/search", handlers.SearchUsers)
			// ----------FRIENDS-------------------------
			protected.GET("/friends", handlers.GetFriends)
			protected.POST("/friends/add", handlers.AddFriend)
			protected.DELETE("/friends/remove", handlers.RemoveFriend)
			// --------ROOMS-------------------------
			protected.GET("/rooms", handlers.GetRooms)
			protected.POST("/rooms/create", handlers.CreateRoom)
			protected.POST("/rooms/add-user", handlers.AddUserToRoom)
			protected.GET("/rooms/:id/members", handlers.GetRoomMembers)
			protected.DELETE("/rooms/:id", handlers.DeleteRoom)
			protected.PUT("/rooms/:id", handlers.UpdateRoom)
			protected.GET("/rooms/name", handlers.GetRoomByName)
			protected.POST("/rooms/invite", handlers.InviteUserToRoom)
			protected.GET("/rooms/invited", handlers.GetInvitedRooms)
			// protected.GET("/rooms/:id/exists", handlers.RoomExists)
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
