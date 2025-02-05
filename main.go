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

	publicAPI := router.Group("/api")
	{
		// Public routes
		publicAPI.POST("/auth/login", handlers.Login)
		publicAPI.POST("/auth/register", handlers.Register)
		publicAPI.POST("/auth/refresh", utils.RefreshToken)
		publicAPI.POST("/auth/validate", utils.ValidateToken)

		// Public check if a room exists
		publicAPI.GET("/rooms/:id/exists", handlers.RoomExists)

		// Public route to list all public rooms
		publicAPI.GET("/rooms/public", handlers.GetAllPublicRooms)

		// Public route to get room info if it's public
		publicAPI.GET("/rooms/:id", handlers.GetRoomByID)

		// Public route to ping-pong
		publicAPI.GET("/ping", utils.PingPong)

		// With auth
		protected := publicAPI.Group("/")
		protected.Use(utils.JWTMiddleware())
		{
			// Users
			protected.GET("/user/id", handlers.GetUserID)
			protected.GET("/friends/search", handlers.SearchUsers)
			protected.GET("/user/:uuid", handlers.GetUserByUUID)
			protected.GET("/user/me", handlers.GetUserByToken)
			
			// Friends
			protected.GET("/friends", handlers.GetFriends)
			protected.POST("/friends/add", handlers.AddFriend)
			protected.DELETE("/friends/remove", handlers.RemoveFriend)
			protected.POST("/friends/request", handlers.RequestFriend)
			protected.POST("/friends/accept", handlers.AcceptFriendRequest)
			protected.POST("/friends/decline", handlers.DeclineFriendRequest)
			protected.GET("/friends/requests", handlers.GetFriendRequests)
			
			// Rooms
			protected.GET("/rooms/mine", handlers.GetMyRooms)
			protected.POST("/rooms/create", handlers.CreateRoom)
			protected.PUT("/rooms/:id", handlers.UpdateRoom)
			protected.DELETE("/rooms/:id", handlers.DeleteRoom)
			protected.POST("/rooms/:id/make-admin", handlers.MakeRoomAdmin)

			// Room invites
			protected.POST("/rooms/invite", handlers.InviteUserToRoom)
			protected.POST("/rooms/invite/accept", handlers.AcceptRoomInvite)
			protected.POST("/rooms/invite/decline", handlers.DeclineRoomInvite)
			protected.GET("/rooms/invites", handlers.GetRoomInvites)
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
