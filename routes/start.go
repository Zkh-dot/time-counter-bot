package routes

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"TimeCounterBot/common"
	"TimeCounterBot/db"
	"TimeCounterBot/tg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartCommand(message *tgbotapi.Message) {
	user, err := db.GetUserByID(common.UserID(message.From.ID))
	if err != nil {
		log.Fatal(err)
	}

	msg := tgbotapi.NewMessage(
		int64(user.ChatID),
		"Hi! You are using Andrew's time management bot.\n"+
			"Firstly, tell me the time interval in which you want to receive question about your activity.",
	)
	msg.ReplyMarkup = getStartCommandTimerIntervalsKeyboardMarkup()

	_, err = bot.Bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}
}

func SetTimerMinutesCallback(callback *tgbotapi.CallbackQuery) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}
	var timerMinutes int64
	_, err = fmt.Sscanf(callback.Data, "set_timer_minutes %d", &timerMinutes)
	if err != nil {
		log.Fatal(err)
	}
	user.TimerMinutes = sql.NullInt64{Int64: timerMinutes, Valid: true}
	err = db.UpdateUser(*user)
	if err != nil {
		log.Fatal(err)
	}

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		int64(user.ChatID),
		callback.Message.MessageID,
		fmt.Sprintf(
			"Nice, your interval is %d minutes! Now tell me the hour in UTC to start sending you reminders.",
			timerMinutes,
		),
		getScheduleMorningStartHourKeyboardMarkup(),
	)

	_, err = bot.Bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}
}

func SetScheduleMorningStartHourCallback(callback *tgbotapi.CallbackQuery) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}
	var scheduleMorningStartHour int64
	_, err = fmt.Sscanf(callback.Data, "schedule_morning_start_hour %d", &scheduleMorningStartHour)
	if err != nil {
		log.Fatal(err)
	}
	user.ScheduleMorningStartHour = sql.NullInt64{Int64: scheduleMorningStartHour, Valid: true}
	err = db.UpdateUser(*user)
	if err != nil {
		log.Fatal(err)
	}

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		int64(user.ChatID),
		callback.Message.MessageID,
		fmt.Sprintf(
			"Wonderful, your start hour will be %d:00 UTC! And now tell me the hour"+
				" in UTC to finish sending reminders and send day statistics.",
			scheduleMorningStartHour,
		),
		getScheduleEveningFinishHourKeyboardMarkup(),
	)

	_, err = bot.Bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}
}

func SetScheduleEveningFinishHourCallback(callback *tgbotapi.CallbackQuery) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}
	var scheduleEveningFinishHour int64
	_, err = fmt.Sscanf(callback.Data, "schedule_evening_finish_hour %d", &scheduleEveningFinishHour)
	if err != nil {
		log.Fatal(err)
	}
	user.ScheduleEveningFinishHour = sql.NullInt64{Int64: scheduleEveningFinishHour, Valid: true}
	err = db.UpdateUser(*user)
	if err != nil {
		log.Fatal(err)
	}

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		int64(user.ChatID),
		callback.Message.MessageID,
		fmt.Sprintf(
			"Cool. You will get notifications every %d minutes, from %d:00 UTC to %d:00 UTC.\n"+
				"Now click the button to enable notifications.",
			user.TimerMinutes.Int64,
			user.ScheduleMorningStartHour.Int64,
			user.ScheduleEveningFinishHour.Int64,
		),
		getEnableNotificationsKeyboardMarkup(),
	)

	_, err = bot.Bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}
}

func EnableNotificationsCallback(callback *tgbotapi.CallbackQuery) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}
	user.TimerEnabled = true
	err = db.UpdateUser(*user)
	if err != nil {
		log.Fatal(err)
	}

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		int64(user.ChatID),
		callback.Message.MessageID,
		fmt.Sprintf(
			"You will get notifications every %d minutes, from %d:00 UTC to %d:00 UTC.\n"+
				"Now click the button to start notifications.\nNotifications enabled!",
			user.TimerMinutes.Int64,
			user.ScheduleMorningStartHour.Int64,
			user.ScheduleEveningFinishHour.Int64,
		),
		getDisableNotificationsKeyboardMarkup(),
	)

	_, err = bot.Bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}
}

func DisableNotificationsCallback(callback *tgbotapi.CallbackQuery) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}
	user.TimerEnabled = false
	err = db.UpdateUser(*user)
	if err != nil {
		log.Fatal(err)
	}

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		int64(user.ChatID),
		callback.Message.MessageID,
		fmt.Sprintf(
			"You will get notifications every %d minutes, from %d:00 UTC to %d:00 UTC.\n"+
				"Now click the button to start notifications.\nNotifications disabled!",
			user.TimerMinutes.Int64,
			user.ScheduleMorningStartHour.Int64,
			user.ScheduleEveningFinishHour.Int64,
		),
		getDisableNotificationsKeyboardMarkup(),
	)

	_, err = bot.Bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}
}

func getStartCommandTimerIntervalsKeyboardMarkup() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		append(
			make([]tgbotapi.InlineKeyboardButton, 0),
			tgbotapi.InlineKeyboardButton{Text: "10 minutes", CallbackData: StringPtr("set_timer_minutes 10")},
			tgbotapi.InlineKeyboardButton{Text: "20 minutes", CallbackData: StringPtr("set_timer_minutes 20")},
		),
		append(
			make([]tgbotapi.InlineKeyboardButton, 0),
			tgbotapi.InlineKeyboardButton{Text: "30 minutes", CallbackData: StringPtr("set_timer_minutes 30")},
			tgbotapi.InlineKeyboardButton{Text: "1 hour", CallbackData: StringPtr("set_timer_minutes 60")},
		),
	)
}

func getScheduleMorningStartHourKeyboardMarkup() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		append(
			make([]tgbotapi.InlineKeyboardButton, 0),
			tgbotapi.InlineKeyboardButton{Text: "03:00", CallbackData: StringPtr("schedule_morning_start_hour 3")},
			tgbotapi.InlineKeyboardButton{Text: "04:00", CallbackData: StringPtr("schedule_morning_start_hour 4")},
			tgbotapi.InlineKeyboardButton{Text: "05:00", CallbackData: StringPtr("schedule_morning_start_hour 5")},
			tgbotapi.InlineKeyboardButton{Text: "06:00", CallbackData: StringPtr("schedule_morning_start_hour 6")},
		),
		append(
			make([]tgbotapi.InlineKeyboardButton, 0),
			tgbotapi.InlineKeyboardButton{Text: "07:00", CallbackData: StringPtr("schedule_morning_start_hour 7")},
			tgbotapi.InlineKeyboardButton{Text: "08:00", CallbackData: StringPtr("schedule_morning_start_hour 8")},
			tgbotapi.InlineKeyboardButton{Text: "09:00", CallbackData: StringPtr("schedule_morning_start_hour 9")},
			tgbotapi.InlineKeyboardButton{Text: "10:00", CallbackData: StringPtr("schedule_morning_start_hour 10")},
		),
	)
}

func getScheduleEveningFinishHourKeyboardMarkup() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		append(
			make([]tgbotapi.InlineKeyboardButton, 0),
			tgbotapi.InlineKeyboardButton{Text: "18:00", CallbackData: StringPtr("schedule_evening_finish_hour 18")},
			tgbotapi.InlineKeyboardButton{Text: "19:00", CallbackData: StringPtr("schedule_evening_finish_hour 19")},
			tgbotapi.InlineKeyboardButton{Text: "20:00", CallbackData: StringPtr("schedule_evening_finish_hour 20")},
			tgbotapi.InlineKeyboardButton{Text: "21:00", CallbackData: StringPtr("schedule_evening_finish_hour 21")},
		),
		append(
			make([]tgbotapi.InlineKeyboardButton, 0),
			tgbotapi.InlineKeyboardButton{Text: "22:00", CallbackData: StringPtr("schedule_evening_finish_hour 22")},
			tgbotapi.InlineKeyboardButton{Text: "23:00", CallbackData: StringPtr("schedule_evening_finish_hour 23")},
			tgbotapi.InlineKeyboardButton{Text: "00:00", CallbackData: StringPtr("schedule_evening_finish_hour 0")},
			tgbotapi.InlineKeyboardButton{Text: "01:00", CallbackData: StringPtr("schedule_evening_finish_hour 1")},
		),
	)
}

func getEnableNotificationsKeyboardMarkup() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		append(
			make([]tgbotapi.InlineKeyboardButton, 0),
			tgbotapi.InlineKeyboardButton{Text: "Enable notifications!", CallbackData: StringPtr("enable_notifications")},
		),
	)
}

func getDisableNotificationsKeyboardMarkup() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		append(
			make([]tgbotapi.InlineKeyboardButton, 0),
			tgbotapi.InlineKeyboardButton{Text: "Disable notifications!", CallbackData: StringPtr("disable_notifications")},
		),
	)
}

func OldStartCommand(message *tgbotapi.Message) {
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

	userState.State = common.InCommand
	waitChan := make(chan string)
	common.UserStates[userID] = common.UserState{State: common.InCommand, WaitingChannel: &waitChan}

	user.TimerMinutes = configureTimerInterval(message.Chat.ID, waitChan)
	user.ScheduleMorningStartHour = configureScheduleMorningStartHour(message.Chat.ID, waitChan)
	user.ScheduleEveningFinishHour = configureScheduleEveningFinishHour(message.Chat.ID, waitChan)

	err = db.UpdateUser(*user)
	if err != nil {
		log.Fatal(err)
	}

	close(waitChan)

	common.UserStates[userID] = common.UserState{State: common.Idle, WaitingChannel: nil}

	_, err = bot.Bot.Send(
		tgbotapi.NewMessage(
			message.Chat.ID,
			"Cool. Now tell me /start_notify when you are ready to get notifications.",
		),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func configureTimerInterval(chatID int64, waitChan chan string) sql.NullInt64 {
	msg := tgbotapi.NewMessage(
		chatID,
		`Hi! You are using Andrew's time management bot. 
		Firstly, tell me the time interval in minutes in which you want to receive question about your activity.`,
	)
	forceReply := tgbotapi.ForceReply{ForceReply: true}
	msg.ReplyMarkup = forceReply

	_, err := bot.Bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}

	ans := <-waitChan

	timerMinutes, err := strconv.ParseInt(ans, 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	return sql.NullInt64{Int64: timerMinutes, Valid: true}
}

func configureScheduleMorningStartHour(chatID int64, waitChan chan string) sql.NullInt64 {
	msg := tgbotapi.NewMessage(
		chatID,
		`Nice! Now tell me the hour in UTC to start sending you reminders.`,
	)
	forceReply := tgbotapi.ForceReply{ForceReply: true}
	msg.ReplyMarkup = forceReply

	_, err := bot.Bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}

	ans := <-waitChan

	scheduleMorningStartHour, err := strconv.ParseInt(ans, 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	return sql.NullInt64{Int64: scheduleMorningStartHour, Valid: true}
}

func configureScheduleEveningFinishHour(chatID int64, waitChan chan string) sql.NullInt64 {
	msg := tgbotapi.NewMessage(
		chatID,
		`Wonderful! And now tell me the hour in UTC to finish sending reminders and send day statistics.`,
	)
	forceReply := tgbotapi.ForceReply{ForceReply: true}
	msg.ReplyMarkup = forceReply

	_, err := bot.Bot.Send(msg)
	if err != nil {
		log.Fatal(err)
	}

	ans := <-waitChan

	scheduleEveningFinishHour, err := strconv.ParseInt(ans, 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	return sql.NullInt64{Int64: scheduleEveningFinishHour, Valid: true}
}

func StringPtr(value string) *string {
	return &value
}
