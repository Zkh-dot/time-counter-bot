package routes

import (
	"database/sql"
	"log"
	"strconv"

	"TimeCounterBot/common"
	"TimeCounterBot/db"
	"TimeCounterBot/tg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartCommand(message *tgbotapi.Message) {
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

	db.UpdateUser(*user)

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
