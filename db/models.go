package db

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// User represents a user in the system
type User struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       string    `gorm:"unique;not null"`
	Username     string    `gorm:"unique;not null"`
	PasswordHash string    `gorm:"not null"`
	Name         string    `gorm:"type:text"`
	Email        string    `gorm:"type:text"`
	IsOnline     bool      `gorm:"default:false"` // Stub for online status
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

// Friend represents a friendship between two users
type Friend struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    string     `gorm:"not null"`
	FriendID  string     `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type FriendRequest struct {
	ID         uint      `gorm:"primaryKey"`
	FromUserID string    `gorm:"not null"`
	ToUserID   string    `gorm:"not null"`
	Status     string    `gorm:"default:'pending';not null"` // pending, accepted, declined
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// Room represents a room
type Room struct {
	ID        uint      `gorm:"primaryKey"`
	RoomID    string    `gorm:"unique;not null"` // UUID
	UserID    string    `gorm:"not null"`       // Creator's user UUID
	Name      string    `gorm:"not null"`
	Type      string    `gorm:"not null"`       // public, private, secret
	Password  string    `gorm:"type:text"`      // null if not password-protected
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// RoomMember represents a member in a room
type RoomMember struct {
	ID        uint      `gorm:"primaryKey"`
	RoomID    string    `gorm:"not null"`       // Room's UUID
	UserID    string    `gorm:"not null"`       // User's UUID
	Role      string    `gorm:"not null"`       // Role in the room (admin, member, viewer)
	JoinedAt  time.Time `gorm:"autoCreateTime"`
}

// For inviting registered users to a room (pending, accepted, declined)
type RoomInvite struct {
	ID            uint      `gorm:"primaryKey"`
	RoomID        string    `gorm:"not null"`
	InviterUserID string    `gorm:"not null"`
	InvitedUserID string    `gorm:"not null"`
	Status        string    `gorm:"default:'pending';not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}

// InitDatabase initializes the SQLite database using GORM
func InitDatabase(path string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate all models
	err = DB.AutoMigrate(
		&User{},
		&Friend{},
		&FriendRequest{},
		&Room{},
		&RoomMember{},
		&RoomInvite{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database schema:", err)
	}
}

// BeforeCreate hook to generate UUID for Room
func (r *Room) BeforeCreate(tx *gorm.DB) (err error) {
	r.RoomID = uuid.New().String()
	return
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.UserID = uuid.New().String()
	return
}
