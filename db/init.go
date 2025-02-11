package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB — глобальный объект для работы с базой через GORM.
var DB *gorm.DB

// InitDB инициализирует подключение к PostgreSQL и выполняет миграции.
func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://bot:secret@localhost:5432/botdb?sslmode=disable"
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("PostgreSQL connection error:", err)
	}

	fmt.Println("✅ Successfully connected to PostgreSQL via GORM")

	// Автоматически создаем/обновляем таблицы для моделей.
	err = DB.AutoMigrate(&Activity{}, &ActivityLog{}, &User{})
	if err != nil {
		log.Fatal("Migration error:", err)
	}
}
