package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Connect() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./tickets.db"
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	// Enable foreign keys
	if _, err = DB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatalf("failed to enable foreign keys: %v", err)
	}

	// WAL mode for better concurrent read performance
	if _, err = DB.Exec("PRAGMA journal_mode = WAL"); err != nil {
		log.Fatalf("failed to set WAL mode: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	fmt.Println("connected to SQLite:", dbPath)
	migrate()
}

func migrate() {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id            TEXT PRIMARY KEY,
		username      TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at    TEXT NOT NULL
	);`

	ticketsTable := `
	CREATE TABLE IF NOT EXISTS tickets (
		id          TEXT PRIMARY KEY,
		user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title       TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		status      TEXT NOT NULL DEFAULT 'open',
		created_at  TEXT NOT NULL
	);`

	for _, q := range []string{usersTable, ticketsTable} {
		if _, err := DB.Exec(q); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
	}

	fmt.Println("database migration completed")
}
