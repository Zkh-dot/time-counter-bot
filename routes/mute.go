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

func MuteActivityCommand(message *tgbotapi.Message, mute bool) {
	tgUser := message.From
	if tgUser == nil {
		return
	}

	userID := common.UserID(tgUser.ID)

	user, err := db.GetUserByID(userID)
	if err != nil {
		log.Fatal(err)
	}

	msgText := "Что хочешь размьютить?"
	callbackCommand := "unmute_activity__unmute"
	var isMuted *bool = nil
	hasMutedLeaves := BoolPtr(true)
	if mute {
		msgText = "Что хочешь замьютить?"
		callbackCommand = "mute_activity__mute"
		isMuted = BoolPtr(false)
		hasMutedLeaves = nil
	}

	msgconf := tgbotapi.NewMessage(int64(user.ChatID), msgText)
	msgconf.ReplyMarkup = buildActivitiesKeyboardMarkupForUser(
		*user, -1, isMuted, hasMutedLeaves, callbackCommand, getMuteActivitiesLastRow(mute))

	_, err = bot.Bot.Send(msgconf)
	if err != nil {
		log.Fatal(err)
	}
}

func MuteActivityCancelCallback(callback *tgbotapi.CallbackQuery) {
	_, err := bot.Bot.Request(
		tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func MuteActivityRefreshCallback(callback *tgbotapi.CallbackQuery, mute bool) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}
	var isMuted *bool = nil
	hasMutedLeaves := BoolPtr(true)
	msgText := "Что хочешь размьютить?"
	callbackCommand := "unmute_activity__unmute"
	if mute {
		isMuted = BoolPtr(false)
		hasMutedLeaves = nil
		msgText = "Что хочешь замьютить?"
		callbackCommand = "mute_activity__mute"
	}

	msgconf := tgbotapi.NewEditMessageTextAndMarkup(
		int64(user.ChatID),
		callback.Message.MessageID,
		msgText,
		buildActivitiesKeyboardMarkupForUser(
			*user, -1, isMuted, hasMutedLeaves, callbackCommand, getMuteActivitiesLastRow(mute)))
	_, err = bot.Bot.Send(msgconf)
	if err != nil {
		log.Fatal(err)
	}
}

func MuteActivityCallback(callback *tgbotapi.CallbackQuery, mute bool) {
	var nodeID int64
	var timerMinutes int64
	var callbackCommand string
	_, err := fmt.Sscanf(callback.Data, "%s %d %d", &callbackCommand, &nodeID, &timerMinutes)
	if err != nil {
		log.Fatal(err)
	}

	var isMuted *bool = nil
	hasMutedLeaves := BoolPtr(true)
	finalMsgFirstPart := "Unmuted activity"
	if mute {
		isMuted = BoolPtr(false)
		hasMutedLeaves = nil
		finalMsgFirstPart = "Muted activity"
	}

	activities, err := db.GetSimpleActivities(common.UserID(callback.From.ID), isMuted, hasMutedLeaves)
	if err != nil {
		log.Fatal(err)
	}

	idx := slices.IndexFunc(activities, func(a db.Activity) bool { return a.ID == nodeID })
	if idx == -1 {
		log.Fatalf("activity with id %d was not found in user activities.", nodeID)
	}

	if activities[idx].IsLeaf {
		if mute {
			err = db.MuteActivityAndMaybeParents(nodeID)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err = db.UnmuteActivityAndMaybeParents(nodeID)
			if err != nil {
				log.Fatal(err)
			}
		}

		activityName, err := db.GetFullActivityNameByID(nodeID, common.UserID(callback.From.ID))
		if err != nil {
			log.Fatal(err)
		}

		_, err = bot.Bot.Send(
			tgbotapi.NewEditMessageTextAndMarkup(
				callback.Message.Chat.ID, callback.Message.MessageID,
				finalMsgFirstPart+" \""+activityName+"\"",
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

		keyboard := buildActivitiesKeyboardMarkupForUser(
			*user, nodeID, isMuted, hasMutedLeaves, callbackCommand, getMuteActivitiesLastRow(mute))

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

func getMuteActivitiesLastRow(mute bool) []tgbotapi.InlineKeyboardButton {
	cancelMuteCallbackText := "unmute_activity__cancel"
	refreshActivitiesCallbackText := "unmute_activity__refresh"
	if mute {
		cancelMuteCallbackText = "mute_activity__cancel"
		refreshActivitiesCallbackText = "mute_activity__refresh"
	}
	return append(
		make([]tgbotapi.InlineKeyboardButton, 0),
		tgbotapi.InlineKeyboardButton{
			Text:         "\U0000274C Cancel",
			CallbackData: &cancelMuteCallbackText,
		},
		tgbotapi.InlineKeyboardButton{
			Text:         "\U0001F504 Refresh activities",
			CallbackData: &refreshActivitiesCallbackText,
		},
	)
}
