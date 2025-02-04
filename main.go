package main

import (
	"context"
	"log"
	"os"

	"TimeCounterBot/db"
	"TimeCounterBot/routes"
	"TimeCounterBot/tg/bot"
	"TimeCounterBot/tg/router"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v3"
)

type conf struct {
	TgToken string `yaml:"telegram_token"`
}

func (c *conf) getConf() *conf {
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

func main() {
	db.InitDB()

	var err error

	var c conf
	c.getConf()

	bot.Bot, err = tgbotapi.NewBotAPI(c.TgToken)
	if err != nil {
		// Abort if something is wrong
		log.Panic(err)
	}

	// Set this to true to log all interactions with telegram servers
	bot.Bot.Debug = false

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	// Create a new cancellable background context. Calling `cancel()` leads to the cancellation of the context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// `updates` is a golang channel which receives telegram updates
	updates := bot.Bot.GetUpdatesChan(updateConfig)

	// Pass cancellable context to goroutine
	go router.ReceiveUpdates(ctx, updates)
	go routes.DispatchNotifications()

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press enter to stop")

	// Wait for a newline symbol, then cancel handling updates
	// _, err = bufio.NewReader(os.Stdin).ReadBytes('\n')
	// if err != nil {
	// 	log.Fatal(err)
	// }

	cancel()
}
