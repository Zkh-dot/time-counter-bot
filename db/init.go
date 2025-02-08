package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL драйвер
)

func InitDB() {
	database := getPostgreSQLDatabase()

	createTableSQL := `CREATE TABLE IF NOT EXISTS activities (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		parent_activity_id INTEGER NOT NULL,
		is_leaf BOOLEAN NOT NULL
	);`
	_, err := database.Exec(createTableSQL)
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

func getPostgreSQLDatabase() *sql.DB {
	if db == nil || db.Ping() != nil {
		// Берем строку подключения из переменной окружения (docker-compose)
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			dsn = "postgres://bot:secret@localhost:5432/botdb?sslmode=disable"
		}

		// Подключаемся к PostgreSQL
		var err error
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			log.Fatal("PostgreSQL connection error:", err)
		}

		// Проверяем соединение
		if err := db.Ping(); err != nil {
			log.Fatal("БД недоступна:", err)
		}
		fmt.Println("✅ Successful connected to PostgreSQL")
	}
	return db
}
