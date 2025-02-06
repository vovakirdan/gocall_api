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
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       string    `gorm:"unique;not null" json:"user_id"`
	Username     string    `gorm:"unique;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"password_hash"`
	Name         string    `gorm:"type:text" json:"name"`
	Email        string    `gorm:"type:text" json:"email"`
	IsOnline     bool      `gorm:"default:false" json:"is_online"` // Stub for online status
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Friend represents a friendship between two users
type Friend struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string     `gorm:"not null" json:"user_id"`
	FriendID  string     `gorm:"not null" json:"friend_id"`
	IsPinned  bool      `gorm:"default:false" json:"is_pinned"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type FriendRequest struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FromUserID string    `gorm:"not null" json:"from_user_id"`
	ToUserID   string    `gorm:"not null" json:"to_user_id"`
	Status     string    `gorm:"default:'pending';not null" json:"status"` // pending, accepted, declined
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Room represents a room
type Room struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RoomID    string    `gorm:"unique;not null" json:"room_id"` // UUID
	UserID    string    `gorm:"not null" json:"user_id"`       // Creator's user UUID
	Name      string    `gorm:"not null" json:"name"`
	Type      string    `gorm:"not null" json:"type"`       // public, private, secret
	Password  string    `gorm:"type:text" json:"password"`      // null if not password-protected
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// RoomMember represents a member in a room
type RoomMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RoomID    string    `gorm:"not null" json:"room_id"`       // Room's UUID
	UserID    string    `gorm:"not null" json:"user_id"`       // User's UUID
	Role      string    `gorm:"not null" json:"role"`       // Role in the room (admin, member, viewer)
	JoinedAt  time.Time `gorm:"autoCreateTime" json:"joined_at"`
}

// For inviting registered users to a room (pending, accepted, declined)
type RoomInvite struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	RoomID        string    `gorm:"not null" json:"room_id"`
	InviterUserID string    `gorm:"not null" json:"inviter_user_id"`
	InvitedUserID string    `gorm:"not null" json:"invited_user_id"`
	Status        string    `gorm:"default:'pending';not null" json:"status"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Message хранит историю текстовых сообщений
type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SenderID   string    `gorm:"not null" json:"sender_id"`   // UUID отправителя
	ReceiverID string    `gorm:"not null" json:"receiver_id"` // UUID получателя
	Text       string    `gorm:"type:text" json:"text"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
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
		&Message{},
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
