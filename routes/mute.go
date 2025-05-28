package routes

import (
	"TimeCounterBot/common"
	"TimeCounterBot/db"
	"TimeCounterBot/tg/bot"
	"fmt"
	"log"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func MuteActivityCommand(message *tgbotapi.Message) {
	tgUser := message.From
	if tgUser == nil {
		return
	}

	userID := common.UserID(tgUser.ID)

	user, err := db.GetUserByID(userID)
	if err != nil {
		log.Fatal(err)
	}

	msgconf := tgbotapi.NewMessage(int64(user.ChatID), "Что хочешь замьютить?")
	isMuted := false
	msgconf.ReplyMarkup = buildActivitiesKeyboardMarkupForUser(
		*user, -1, &isMuted, "mute_activity__mute", getMuteActivitiesLastRow())

	_, err = bot.Bot.Send(msgconf)
	if err != nil {
		log.Fatal(err)
	}
}

func MuteActivityCancelCallback(callback *tgbotapi.CallbackQuery) {
	_, err := bot.Bot.Send(
		tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func MuteActivityRefreshCallback(callback *tgbotapi.CallbackQuery) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}
	isMuted := false
	msgconf := tgbotapi.NewEditMessageTextAndMarkup(
		int64(user.ChatID),
		callback.Message.MessageID,
		"Что хочешь замьютить?",
		buildActivitiesKeyboardMarkupForUser(
			*user, -1, &isMuted, "mute_activity__mute", getMuteActivitiesLastRow()))
	_, err = bot.Bot.Send(msgconf)
	if err != nil {
		log.Fatal(err)
	}
}

func MuteActivityCallback(callback *tgbotapi.CallbackQuery) {
	var nodeID int64

	var timerMinutes int64

	_, err := fmt.Sscanf(callback.Data, "mute_activity__mute %d %d", &nodeID, &timerMinutes)
	if err != nil {
		log.Fatal(err)
	}

	isMuted := false
	activities, err := db.GetSimpleActivities(common.UserID(callback.From.ID), &isMuted)
	if err != nil {
		log.Fatal(err)
	}

	idx := slices.IndexFunc(activities, func(a db.Activity) bool { return a.ID == nodeID })
	if idx == -1 {
		log.Fatalf("activity with id %d was not found in user activities.", nodeID)
	}

	if activities[idx].IsLeaf {
		db.SetIsMutedActivity(activities[idx].ID, true)

		activityName, err := db.GetFullActivityNameByID(nodeID, common.UserID(callback.From.ID))
		if err != nil {
			log.Fatal(err)
		}

		_, err = bot.Bot.Send(
			tgbotapi.NewEditMessageTextAndMarkup(
				callback.Message.Chat.ID, callback.Message.MessageID,
				"Muted activity \""+activityName+"\"",
				tgbotapi.InlineKeyboardMarkup{InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0)},
			),
		)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		user, err := db.GetUserByID(common.UserID(callback.From.ID))
		if err != nil {
			log.Fatal(err)
		}

		isMuted := false
		keyboard := buildActivitiesKeyboardMarkupForUser(
			*user, nodeID, &isMuted, "mute_activity__mute", getMuteActivitiesLastRow())

		_, err = bot.Bot.Send(
			tgbotapi.NewEditMessageTextAndMarkup(
				callback.Message.Chat.ID, callback.Message.MessageID, callback.Message.Text, keyboard,
			),
		)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getMuteActivitiesLastRow() []tgbotapi.InlineKeyboardButton {
	CancelMuteCallbackText := "mute_activity__cancel"
	refreshActivitiesCallbackText := "mute_activity__refresh"
	return append(
		make([]tgbotapi.InlineKeyboardButton, 0),
		tgbotapi.InlineKeyboardButton{
			Text:         "\U0000274C Cancel",
			CallbackData: &CancelMuteCallbackText,
		},
		tgbotapi.InlineKeyboardButton{
			Text:         "\U0001F504 Refresh activities",
			CallbackData: &refreshActivitiesCallbackText,
		},
	)
}

func UnmuteActivityCommand(message *tgbotapi.Message) {
}

func UnmuteActivityCallback(callback *tgbotapi.CallbackQuery) {
}
