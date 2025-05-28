package routes

import (
	"TimeCounterBot/common"
	"TimeCounterBot/db"
	tg "TimeCounterBot/tg/bot"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const Day time.Duration = time.Duration(24) * time.Hour
const DayStatsWaitDuration time.Duration = 5 * time.Second

func TestDayStatsRoutine(message *tgbotapi.Message) {
	user, err := db.GetUserByID(common.UserID(message.From.ID))
	if err != nil {
		log.Fatal(err)
	}
	startDayStatsRoutine(*user)
}

func startDayStatsRoutine(user db.User) {
	time.Sleep(DayStatsWaitDuration)
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
	log.Println("buildDayStatsRoutineKeyboardMarkup: ", now)
	callbackData := fmt.Sprintf(
		"day_stats__send_chart %d %d",
		now.Add(-Day).Unix(),
		now.Unix(),
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

	var tsStartUnix int64
	var tsEndUnix int64
	_, err = fmt.Sscanf(callback.Data, "day_stats__send_chart %d %d", &tsStartUnix, &tsEndUnix)
	if err != nil {
		log.Fatal(err)
	}
	tsStart := time.Unix(tsStartUnix, 0)
	tsEnd := time.Unix(tsEndUnix, 0)

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
			tgbotapi.NewInlineKeyboardButtonData(
				"\U0001F504 Refresh chart",
				fmt.Sprintf("day_stats__refresh_chart %d %d", tsStartUnix, tsEndUnix)),
		),
	)

	_, err = tg.Bot.Send(msgconf)
	if err != nil {
		log.Fatalf("❌ Ошибка отправки изображения: %v", err)
	}

	_, err = tg.Bot.Send(tgbotapi.NewDeleteMessage(int64(user.ChatID), callback.Message.MessageID))
	if err != nil && !strings.Contains(err.Error(), "cannot unmarshal bool into Go value of type tgbotapi.Message") {
		log.Fatal(err)
	}

	err = os.Remove(outputFile)
	if err != nil {
		log.Println("Couldn't delete output file: ", err)
	}
}

func RefreshDayStatsChartCallback(callback *tgbotapi.CallbackQuery) {
	user, err := db.GetUserByID(common.UserID(callback.From.ID))
	if err != nil {
		log.Fatal(err)
	}

	var tsStartUnix int64
	var tsEndUnix int64
	_, err = fmt.Sscanf(callback.Data, "day_stats__refresh_chart %d %d", &tsStartUnix, &tsEndUnix)
	if err != nil {
		log.Fatal(err)
	}
	tsStart := time.Unix(tsStartUnix, 0)
	tsEnd := time.Unix(tsEndUnix, 0)

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
				fmt.Sprintf("day_stats__refresh_chart %d %d", tsStartUnix, tsEndUnix)),
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

	err = os.Remove(outputFile)
	if err != nil {
		log.Println("Couldn't delete output file: ", err)
	}
}
