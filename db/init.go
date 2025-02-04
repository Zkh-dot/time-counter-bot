package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() {
	mutex.Lock()
	defer mutex.Unlock()

	database, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS activities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		parent_activity_id INTEGER NOT NULL,
		is_leaf BOOLEAN NOT NULL
	);`
	_, err = database.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL = `CREATE TABLE IF NOT EXISTS activity_log (
		message_id INTEGER,
		user_id INTEGER,
		activity_id INTEGER NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		interval_minutes INTEGER NOT NULL,

		PRIMARY KEY (message_id, user_id)
	);`
	_, err = database.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL = `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		chat_id INTEGER,
		timer_enabled BOOLEAN NOT NULL,
		timer_minutes INTEGER,
		schedule_morning_start_hour INTEGER,
		schedule_evening_finish_hour INTEGER,
		last_notify TIMESTAMP
	);`
	_, err = database.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}
