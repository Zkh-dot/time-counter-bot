package routes

import (
	"TimeCounterBot/common"
	"TimeCounterBot/db"
	tg "TimeCounterBot/tg/bot"
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestDayStatsRoutine(message *tgbotapi.Message) {
	user, err := db.GetUserByID(common.UserID(message.From.ID))
	if err != nil {
		log.Fatal(err)
	}
	startDayStatsRoutine(*user)
}

func startDayStatsRoutine(user db.User) {
	msgconf := tgbotapi.NewMessage(int64(user.ChatID), "Если заполнил все активности за сегодня - ЖМИ НА КНОПКУ!")
	msgconf.ReplyMarkup = buildDayStatsRoutineKeyboardMarkup()

	_, err := tg.Bot.Send(msgconf)
	if err != nil {
		log.Fatal(err)
	}
}

func buildDayStatsRoutineKeyboardMarkup() tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 1)
	rows[0] = make([]tgbotapi.InlineKeyboardButton, 1)

	now := time.Now()
	callbackData := fmt.Sprintf(
		"send_day_stats_chart %s %s", now.Add(-time.Duration(24)*time.Hour).Format(time.RFC3339), now.Format(time.RFC3339),
	)
	rows[0][0] = tgbotapi.InlineKeyboardButton{
		Text:         "Кнопка",
		CallbackData: &callbackData,
	}

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func SendDayStatsRoutineCallback(callback *tgbotapi.CallbackQuery) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}

	var tsStartStr string
	var tsEndStr string
	_, err = fmt.Sscanf(callback.Data, "send_day_stats_chart %s %s", &tsStartStr, &tsEndStr)
	if err != nil {
		log.Fatal(err)
	}
	tsStart, err := time.Parse(time.RFC3339, tsStartStr)
	if err != nil {
		log.Fatal(err)
	}
	tsEnd, err := time.Parse(time.RFC3339, tsEndStr)
	if err != nil {
		log.Fatal(err)
	}

	data := getUserActivityDataForInterval(*user, tsStart, tsEnd)

	outputFile := fmt.Sprintf("sunburst_chart_%d_%d.png", user.ID, callback.Message.MessageID)
	generateActivityChart(data, outputFile)

	// Отправляем картинку в Telegram
	msgconf := tgbotapi.NewPhoto(int64(user.ChatID), tgbotapi.FilePath(outputFile))
	msgconf.Caption = fmt.Sprintf(
		"Диаграмма активности за сегодняшний день (сделал в %s за интервал [%s, %s]).",
		time.Now().Format(time.RFC3339), tsStart, tsEnd,
	)

	msgconf.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("\U0001F504 Refresh chart", fmt.Sprintf("refresh_day_stats_chart %s %s", tsStart, tsEnd)),
		),
	)

	_, err = tg.Bot.Send(msgconf)
	if err != nil {
		log.Fatalf("❌ Ошибка отправки изображения: %v", err)
	}

	_, err = tg.Bot.Send(tgbotapi.NewDeleteMessage(int64(user.ChatID), callback.Message.MessageID))
	if err != nil {
		log.Fatal(err)
	}

	os.Remove(outputFile)
}

func RefreshDayStatsChartCallback(callback *tgbotapi.CallbackQuery) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}

	var tsStartStr string
	var tsEndStr string
	_, err = fmt.Sscanf(callback.Data, "refresh_day_stats_chart %s %s", &tsStartStr, &tsEndStr)
	if err != nil {
		log.Fatal(err)
	}
	tsStart, err := time.Parse(time.RFC3339, tsStartStr)
	if err != nil {
		log.Fatal(err)
	}
	tsEnd, err := time.Parse(time.RFC3339, tsEndStr)
	if err != nil {
		log.Fatal(err)
	}

	data := getUserActivityDataForInterval(*user, tsStart, tsEnd)

	outputFile := fmt.Sprintf("sunburst_chart_%d_%d.png", user.ID, callback.Message.MessageID)
	generateActivityChart(data, outputFile)

	// Отправляем картинку в Telegram
	newPhoto := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(outputFile))
	newPhoto.Caption = fmt.Sprintf(
		"Диаграмма активности за сегодняшний день (сделал в %s за интервал [%s, %s]).",
		time.Now().Format(time.RFC3339), tsStart, tsEnd,
	)

	newKeyboardMarkup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"\U0001F504 Refresh chart",
				fmt.Sprintf("refresh_day_stats_chart %s %s", tsStart, tsEnd)),
		),
	)
	// Формируем конфигурацию редактирования медиа
	editMedia := tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      int64(user.ChatID),
			MessageID:   callback.Message.MessageID,
			ReplyMarkup: &newKeyboardMarkup,
		},
		Media: newPhoto,
	}

	_, err = tg.Bot.Send(editMedia)
	if err != nil {
		log.Fatal(err)
	}

	os.Remove(outputFile)
}
