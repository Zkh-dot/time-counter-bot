package routes

import (
	"database/sql"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"TimeCounterBot/common"
	"TimeCounterBot/db"
	tg "TimeCounterBot/tg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// notifyUser -> sends message M and creates keybord Ki with first-level activities
// user sends callback on K1 [activity_log node_id timestamp]
// LogUserActivityCallback gets callback, switch:
//  node_id is a leaf -> logs leaf-activity, deletes Ki
//  node_id is not a leaf -> load all children of node_id, creates new Keyboard Ki+1

func notifyUser(user db.User) {
	user.LastNotify = sql.NullTime{Time: time.Now(), Valid: true}
	db.UpdateUser(user)

	msgconf := tgbotapi.NewMessage(int64(user.ChatID), "Чё делаеш?))0)")
	msgconf.ReplyMarkup = buildActivitiesKeyboardMarkupForUser(user, -1)

	_, err := tg.Bot.Send(msgconf)
	if err != nil {
		log.Fatal(err)
	}
}

func LogUserActivityCallback(callback *tgbotapi.CallbackQuery) {
	var nodeID int64

	var timerMinutes int64

	_, err := fmt.Sscanf(callback.Data, "activity_log %d %d", &nodeID, &timerMinutes)
	if err != nil {
		log.Fatal(err)
	}

	activities := db.GetSimpleActivities(common.UserID(callback.From.ID))

	idx := slices.IndexFunc(activities, func(a db.Activity) bool { return a.ID == nodeID })
	if idx == -1 {
		log.Fatalf("activity with id %d was not found in user activities.", nodeID)
	}

	if activities[idx].IsLeaf {
		db.AddActivityLog(
			db.ActivityLog{
				MessageID:       int64(callback.Message.MessageID),
				UserID:          callback.From.ID,
				ActivityID:      nodeID,
				Timestamp:       callback.Message.Time(),
				IntervalMinutes: timerMinutes,
			},
		)

		activityName, err := db.GetFullActivityNameByID(nodeID, common.UserID(callback.From.ID))
		if err != nil {
			log.Fatal(err)
		}

		_, err = tg.Bot.Send(
			tgbotapi.NewEditMessageTextAndMarkup(
				callback.Message.Chat.ID, callback.Message.MessageID,
				"Saved activity \""+activityName+"\"",
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

		keyboard := buildActivitiesKeyboardMarkupForUser(*user, nodeID)

		_, err = tg.Bot.Send(
			tgbotapi.NewEditMessageTextAndMarkup(
				callback.Message.Chat.ID, callback.Message.MessageID, callback.Message.Text, keyboard,
			),
		)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func RefreshActivitiesCallback(callback *tgbotapi.CallbackQuery) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}

	keyboard := buildActivitiesKeyboardMarkupForUser(*user, -1)

	_, err = tg.Bot.Send(
		tgbotapi.NewEditMessageTextAndMarkup(
			callback.Message.Chat.ID, callback.Message.MessageID, callback.Message.Text, keyboard,
		),
	)
	if err != nil && !strings.Contains(err.Error(), "message is not modified") {
		log.Fatal(err)
	}
}

func AddNewActivityCallback(callback *tgbotapi.CallbackQuery) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}

	registerNewActivity(*user)
}

func buildActivitiesKeyboardMarkupForUser(user db.User, parentActivityID int64) tgbotapi.InlineKeyboardMarkup {
	activities := db.GetSimpleActivities(user.ID)

	rows := make([][]tgbotapi.InlineKeyboardButton, 0)

	for _, activity := range activities {
		if activity.ParentActivityID != parentActivityID {
			continue
		}

		leafIDStr := fmt.Sprintf(
			"activity_log %d %d", activity.ID, user.TimerMinutes.Int64,
		)
		buttons := make([]tgbotapi.InlineKeyboardButton, 0)
		buttons = append(
			buttons,
			tgbotapi.InlineKeyboardButton{
				Text:         activity.Name,
				CallbackData: &leafIDStr,
			},
		)
		rows = append(rows, buttons)
	}

	rows = append(rows, getStandardActivitiesLastRow())

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func getStandardActivitiesLastRow() []tgbotapi.InlineKeyboardButton {
	newActivityCallbackText := "register_new_activity"
	refreshActivitiesCallbackText := "refresh_activities"
	return append(
		make([]tgbotapi.InlineKeyboardButton, 0),
		tgbotapi.InlineKeyboardButton{
			Text:         "Add new activity",
			CallbackData: &newActivityCallbackText,
		},
		tgbotapi.InlineKeyboardButton{
			Text:         "\U0001F504 Refresh activities",
			CallbackData: &refreshActivitiesCallbackText,
		},
	)
}
