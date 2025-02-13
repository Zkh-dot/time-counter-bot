package router

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"TimeCounterBot/common"
	"TimeCounterBot/db"
	"TimeCounterBot/routes"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ReceiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for {
		select {
		// stop looping if ctx is cancelled
		case <-ctx.Done():
			return
		// receive update from channel and then handle it
		case update := <-updates:
			go handleUpdate(update)
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	switch {
	// Handle messages
	case update.Message != nil:
		handleMessage(update.Message)

	case update.CallbackQuery != nil:
		handleCallbackQuery(update.CallbackQuery)
	}
}

func handleMessage(message *tgbotapi.Message) {
	user := message.From
	if user == nil {
		return
	}

	userID := common.UserID(user.ID)

	MaybeAddNewUser(userID, common.ChatID(message.Chat.ID))

	// Print to console
	log.Printf("%s wrote %s", user.FirstName, message.Text)

	if strings.HasPrefix(message.Text, "/") {
		handleCommand(message)
	} else if len(message.Text) > 0 {
		// chech user state and send info to waiting channel
		if common.UserStates[userID].WaitingChannel != nil {
			*common.UserStates[userID].WaitingChannel <- message.Text
		}
	}
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	dataPath := strings.Split(callback.Data, " ")[0]
	switch dataPath {
	case "activity_log":
		routes.LogUserActivityCallback(callback)

	case "register_new_activity":
		routes.AddNewActivityCallback(callback)

	case "refresh_activities":
		routes.RefreshActivitiesCallback(callback)

	case "day_stats__send_chart":
		routes.SendDayStatsRoutineCallback(callback)

	case "day_stats__refresh_chart":
		routes.RefreshDayStatsChartCallback(callback)

	case "start__set_timer_minutes":
		routes.SetTimerMinutesCallback(callback)

	case "start__schedule_morning_start_hour":
		routes.SetScheduleMorningStartHourCallback(callback)

	case "start__schedule_evening_finish_hour":
		routes.SetScheduleEveningFinishHourCallback(callback)

	case "start__enable_notifications":
		routes.EnableNotificationsCallback(callback)

	case "start__disable_notifications":
		routes.DisableNotificationsCallback(callback)
	}
}

// When we get a command, we react accordingly.
func handleCommand(message *tgbotapi.Message) {
	switch strings.Split(message.Text, " ")[0] {
	case "/start":
		routes.StartCommand(message)

	case "/start_notify":
		routes.StartNotifyCommand(message)

	case "/stop_notify":
		routes.StopNotifyCommand(message)

	case "/test_notify":
		routes.TestNotifyCommand(message)

	case "/register_new_activity":
		routes.RegisterNewActivityCommand(message)

	case "/get_day_statistics":
		routes.GetDayStatisticsCommand(message)

	case "/test_day_stats_routine":
		routes.TestDayStatsRoutine(message)

	}
}

func MaybeAddNewUser(userID common.UserID, chatID common.ChatID) {
	_, err := db.GetUserByID(userID)
	if err != nil && !strings.Contains(err.Error(), "was not found") {
		log.Fatal(err)
	}

	if err != nil && strings.Contains(err.Error(), "was not found") {
		err = db.AddUser(
			db.User{
				ID:                        userID,
				ChatID:                    chatID,
				TimerEnabled:              false,
				TimerMinutes:              sql.NullInt64{},
				ScheduleMorningStartHour:  sql.NullInt64{},
				ScheduleEveningFinishHour: sql.NullInt64{},
				LastNotify:                sql.NullTime{},
			},
		)
		if err != nil {
			log.Fatal(err)
		}
	}
}
