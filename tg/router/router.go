package router

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"TimeCounterBot/common"
	"TimeCounterBot/db"
	"TimeCounterBot/routes"
	"TimeCounterBot/tg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SetCommands() {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Начать работу с ботом",
		},
		{
			Command:     "register_new_activity",
			Description: "Зарегистрировать новую активность",
		},
		{
			Command:     "start_notify",
			Description: "Начать присылать уведомления",
		},
		{
			Command:     "stop_notify",
			Description: "Закончить присылать уведомления",
		},
		{
			Command:     "mute_activity",
			Description: "Замьютить активность (чтобы не отображалась в регулярных опросах)",
		},
		{
			Command:     "unmute_activity",
			Description: "Размьютить активность (чтобы снова появилась в регулярных опросах)",
		},
		{
			Command:     "analytics",
			Description: "Аналитика и статистика активностей",
		},
		{
			Command:     "export_activities",
			Description: "Экспортировать дерево активностей в YAML файл",
		},
		{
			Command:     "import_activities",
			Description: "Импортировать активности из YAML файла",
		},
		{
			Command:     "delete_activity",
			Description: "Удалить активность и все её подактивности",
		},
	}

	setCmd := tgbotapi.NewSetMyCommands(commands...)
	_, err := bot.Bot.Request(setCmd)
	if err != nil {
		log.Printf("Ошибка установки команд: %v", err)
	}
}

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

	maybeAddNewUser(userID, common.ChatID(message.Chat.ID))

	// Print to console
	log.Printf("%s wrote %s", user.UserName, message.Text)

	if strings.HasPrefix(message.Text, "/") {
		handleCommand(message)
	} else if message.Document != nil {
		// Обрабатываем загруженный документ (возможно, для импорта активностей)
		routes.ProcessImportFile(message)
	} else if len(message.Text) > 0 {
		// chech user state and send info to waiting channel
		if common.UserStates[userID].WaitingChannel != nil {
			*common.UserStates[userID].WaitingChannel <- message.Text
		}
	}
}

type CallbackHandler func(*tgbotapi.CallbackQuery)

var callbackHandlers = map[string]CallbackHandler{
	"activity_log":          routes.LogUserActivityCallback,
	"register_new_activity": routes.AddNewActivityCallback,
	"refresh_activities":    routes.RefreshActivitiesCallback,

	"day_stats__send_chart":    routes.SendDayStatsRoutineCallback,
	"day_stats__refresh_chart": routes.RefreshDayStatsChartCallback,

	"start__set_timer_minutes":            routes.SetTimerMinutesCallback,
	"start__schedule_morning_start_hour":  routes.SetScheduleMorningStartHourCallback,
	"start__schedule_evening_finish_hour": routes.SetScheduleEveningFinishHourCallback,
	"start__enable_notifications": func(c *tgbotapi.CallbackQuery) {
		routes.EnableNotificationsCallback(c, true)
	},
	"start__disable_notifications": func(c *tgbotapi.CallbackQuery) {
		routes.EnableNotificationsCallback(c, false)
	},

	"mute_activity__mute":    func(c *tgbotapi.CallbackQuery) { routes.MuteActivityCallback(c, true) },
	"mute_activity__cancel":  routes.MuteActivityCancelCallback,
	"mute_activity__refresh": func(c *tgbotapi.CallbackQuery) { routes.MuteActivityRefreshCallback(c, true) },

	"unmute_activity__unmute":  func(c *tgbotapi.CallbackQuery) { routes.MuteActivityCallback(c, false) },
	"unmute_activity__cancel":  routes.MuteActivityCancelCallback,
	"unmute_activity__refresh": func(c *tgbotapi.CallbackQuery) { routes.MuteActivityRefreshCallback(c, false) },

	"delete_activity__delete":  routes.DeleteActivityCallback,
	"delete_activity__cancel":  routes.DeleteActivityCancelCallback,
	"delete_activity__refresh": routes.DeleteActivityRefreshCallback,

	"analytics__day_stats":       routes.AnalyticsGetDayStatsCallback,
	"analytics__compare_periods": routes.AnalyticsComperiodsCallback,
	"analytics__back":            routes.AnalyticsBackCallback,

	"compare_periods__this_vs_last_week":  routes.ComparePeriods_ThisVsLastWeekCallback,
	"compare_periods__this_vs_last_month": routes.ComparePeriods_ThisVsLastMonthCallback,
	"compare_periods__custom":             routes.ComparePeriods_CustomCallback,
	"compare_periods__back":               routes.ComparePeriods_BackCallback,

	"day_stats__today":     func(c *tgbotapi.CallbackQuery) { routes.DayStatsCallback(c, "today") },
	"day_stats__yesterday": func(c *tgbotapi.CallbackQuery) { routes.DayStatsCallback(c, "yesterday") },
	"day_stats__this_week": func(c *tgbotapi.CallbackQuery) { routes.DayStatsCallback(c, "this_week") },
	"day_stats__last_week": func(c *tgbotapi.CallbackQuery) { routes.DayStatsCallback(c, "last_week") },
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	dataPath := strings.Split(callback.Data, " ")[0]
	if handler, ok := callbackHandlers[dataPath]; ok {
		handler(callback)
	} else {
		log.Printf("Unknown callback: %q", dataPath)
	}
}

// When we get a command, we react accordingly.
func handleCommand(message *tgbotapi.Message) {
	switch strings.Split(message.Text, " ")[0] {
	case "/start":
		routes.StartCommand(message)

	case "/start_notify":
		routes.NotifyCommand(message, true)

	case "/stop_notify":
		routes.NotifyCommand(message, false)

	case "/test_notify":
		routes.TestNotifyCommand(message)

	case "/register_new_activity":
		routes.RegisterNewActivityCommand(message)

	case "/analytics":
		routes.AnalyticsMenuCommand(message)

	case "/get_day_statistics":
		routes.GetDayStatisticsCommand(message)

	case "/test_day_stats_routine":
		routes.TestDayStatsRoutine(message)

	case "/mute_activity":
		routes.MuteActivityCommand(message, true)

	case "/unmute_activity":
		routes.MuteActivityCommand(message, false)

	case "/export_activities":
		routes.ExportActivitiesCommand(message)

	case "/import_activities":
		routes.ImportActivitiesCommand(message)

	case "/delete_activity":
		routes.DeleteActivityCommand(message)
	}
}

func maybeAddNewUser(userID common.UserID, chatID common.ChatID) {
	_, err := db.GetUserByID(userID)
	if err != nil && !strings.Contains(err.Error(), "record not found") {
		log.Fatal(err)
	}

	if err != nil && strings.Contains(err.Error(), "record not found") {
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
