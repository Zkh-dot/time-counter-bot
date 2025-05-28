package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"TimeCounterBot/db"
	"TimeCounterBot/routes"
	"TimeCounterBot/tg/bot"
	"TimeCounterBot/tg/router"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	db.InitDB()

	var err error

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("Telegram token was not found")
	}

	bot.Bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		// Abort if something is wrong
		log.Panic(err)
	}

	// Set this to true to log all interactions with telegram servers
	bot.Bot.Debug = false

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	// Создаем контекст с возможностью отмены.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // гарантируем вызов cancel при завершении main

	// Настраиваем обработку сигналов для корректного завершения работы.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Получен сигнал %s, завершаем работу...", sig)
		cancel()
	}()

	updates := bot.Bot.GetUpdatesChan(updateConfig)

	go router.SetCommands()
	go router.ReceiveUpdates(ctx, updates)
	go routes.DispatchNotifications()

	log.Println("Start listening for updates. Press enter to stop")

	// Блокируем выполнение до получения cancel.
	<-ctx.Done()
	log.Println("Shutting down...")
}
