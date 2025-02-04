package routes

import (
	"log"

	"TimeCounterBot/common"
	"TimeCounterBot/db"
	"TimeCounterBot/tg/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartNotifyCommand(message *tgbotapi.Message) {
	tgUser := message.From
	if tgUser == nil {
		return
	}

	userID := common.UserID(tgUser.ID)

	user, err := db.GetUserByID(userID)
	if err != nil {
		log.Fatal(err)
	}

	userState := common.UserStates[userID]

	if userState.State == common.InCommand {
		_, err = bot.Bot.Send(
			tgbotapi.NewMessage(message.Chat.ID, "You're already executing some command"),
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	user.TimerEnabled = true
	db.UpdateUser(*user)
}

func StopNotifyCommand(message *tgbotapi.Message) {
	tgUser := message.From
	if tgUser == nil {
		return
	}

	userID := common.UserID(tgUser.ID)

	user, err := db.GetUserByID(userID)
	if err != nil {
		log.Fatal(err)
	}

	userState := common.UserStates[userID]

	if userState.State == common.InCommand {
		_, err = bot.Bot.Send(
			tgbotapi.NewMessage(message.Chat.ID, "You're already executing some command"),
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	user.TimerEnabled = false
	db.UpdateUser(*user)
}

func TestNotifyCommand(message *tgbotapi.Message) {
	tgUser := message.From
	if tgUser == nil {
		return
	}

	userID := common.UserID(tgUser.ID)

	user, err := db.GetUserByID(userID)
	if err != nil {
		log.Fatal(err)
	}

	userState := common.UserStates[userID]

	if userState.State == common.InCommand {
		_, err = bot.Bot.Send(
			tgbotapi.NewMessage(message.Chat.ID, "You're already executing some command"),
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	notifyUser(*user)
}
