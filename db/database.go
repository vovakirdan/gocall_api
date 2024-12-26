package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDatabase(path string) {
	var err error
	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	createTables()
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS friends (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			friend_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS rooms (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS room_members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			room_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			joined_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			log.Fatal("Failed to execute query:", err)
		}
	}
}
