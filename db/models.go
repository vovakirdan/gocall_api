package db

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// User represents a user in the system
type User struct {
	ID uint `gorm:"primaryKey"`
	Username string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	Name string `gorm:"type:text"`  // may be null
	Email string `gorm:"type:text"` // may be null
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// Friend represents a friendship between two users
type Friend struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	FriendID  uint      `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// Room represents a room
type Room struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	Name      string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// RoomMember represents a member in a room
type RoomMember struct {
	ID        uint      `gorm:"primaryKey"`
	RoomID    uint      `gorm:"not null"`
	UserID    uint      `gorm:"not null"`
	JoinedAt  time.Time `gorm:"autoCreateTime"`
}

// InitDatabase initializes the SQLite database using GORM
func InitDatabase(path string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate all models
	err = DB.AutoMigrate(&User{}, &Friend{}, &Room{}, &RoomMember{})
	if err != nil {
		log.Fatal("Failed to migrate database schema:", err)
	}
}