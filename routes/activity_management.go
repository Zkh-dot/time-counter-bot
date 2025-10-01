package routes

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"TimeCounterBot/common"
	"TimeCounterBot/db"
	"TimeCounterBot/tg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ExportActivitiesCommand обрабатывает команду экспорта активностей.
func ExportActivitiesCommand(message *tgbotapi.Message) {
	tgUser := message.From
	if tgUser == nil {
		return
	}

	userID := common.UserID(tgUser.ID)

	user, err := db.GetUserByID(userID)
	if err != nil {
		log.Printf("Ошибка получения пользователя: %v", err)
		return
	}

	// Экспортируем активности в YAML
	yamlData, err := db.ExportActivitiesToYAML(userID)
	if err != nil {
		log.Printf("Ошибка экспорта активностей: %v", err)
		msgConf := tgbotapi.NewMessage(int64(user.ChatID), "Произошла ошибка при экспорте активностей.")
		bot.Bot.Send(msgConf)
		return
	}

	// Создаем документ для отправки
	document := tgbotapi.NewDocument(int64(user.ChatID), tgbotapi.FileBytes{
		Name:  fmt.Sprintf("activities_export_%d.yaml", userID),
		Bytes: yamlData,
	})
	document.Caption = "Экспорт ваших активностей в формате YAML"

	_, err = bot.Bot.Send(document)
	if err != nil {
		log.Printf("Ошибка отправки файла: %v", err)
		msgConf := tgbotapi.NewMessage(int64(user.ChatID), "Произошла ошибка при отправке файла.")
		bot.Bot.Send(msgConf)
		return
	}

	// Удаляем исходное сообщение команды
	_, err = bot.Bot.Request(
		tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID),
	)
	if err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}
}

// ImportActivitiesCommand обрабатывает команду импорта активностей.
func ImportActivitiesCommand(message *tgbotapi.Message) {
	tgUser := message.From
	if tgUser == nil {
		return
	}

	userID := common.UserID(tgUser.ID)

	user, err := db.GetUserByID(userID)
	if err != nil {
		log.Printf("Ошибка получения пользователя: %v", err)
		return
	}

	// Устанавливаем состояние ожидания файла
	state := common.UserStates[userID]
	if state.WaitingChannel == nil {
		waitingChan := make(chan string, 1)
		common.UserStates[userID] = common.UserState{
			WaitingChannel: &waitingChan,
		}
	}

	msgConf := tgbotapi.NewMessage(int64(user.ChatID),
		"Пришлите YAML файл с экспортированными активностями для импорта.\n\n"+
			"⚠️ Внимание: импорт добавит новые активности к существующим, не заменяя их полностью.")

	_, err = bot.Bot.Send(msgConf)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
		return
	}

	// Ждем файл от пользователя
	go func() {
		select {
		case <-*common.UserStates[userID].WaitingChannel:
			// Пользователь отправил что-то, но нам нужен именно документ
			msgConf := tgbotapi.NewMessage(int64(user.ChatID),
				"Пожалуйста, отправьте YAML файл как документ, а не текст.")
			bot.Bot.Send(msgConf)
		}
	}()

	// Удаляем исходное сообщение команды
	_, err = bot.Bot.Request(
		tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID),
	)
	if err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}
}

// ProcessImportFile обрабатывает загруженный YAML файл для импорта.
func ProcessImportFile(message *tgbotapi.Message) {
	if message.Document == nil {
		return
	}

	tgUser := message.From
	if tgUser == nil {
		return
	}

	userID := common.UserID(tgUser.ID)

	user, err := db.GetUserByID(userID)
	if err != nil {
		log.Printf("Ошибка получения пользователя: %v", err)
		return
	}

	// Проверяем расширение файла
	if !strings.HasSuffix(strings.ToLower(message.Document.FileName), ".yaml") &&
		!strings.HasSuffix(strings.ToLower(message.Document.FileName), ".yml") {
		msgConf := tgbotapi.NewMessage(int64(user.ChatID),
			"Поддерживаются только YAML файлы (.yaml или .yml)")
		bot.Bot.Send(msgConf)
		return
	}

	// Получаем файл
	fileConfig := tgbotapi.FileConfig{FileID: message.Document.FileID}
	file, err := bot.Bot.GetFile(fileConfig)
	if err != nil {
		log.Printf("Ошибка получения файла: %v", err)
		msgConf := tgbotapi.NewMessage(int64(user.ChatID), "Ошибка загрузки файла.")
		bot.Bot.Send(msgConf)
		return
	}

	// Скачиваем содержимое файла
	resp, err := http.Get(file.Link(bot.Bot.Token))
	if err != nil {
		log.Printf("Ошибка скачивания файла: %v", err)
		msgConf := tgbotapi.NewMessage(int64(user.ChatID), "Ошибка скачивания файла.")
		bot.Bot.Send(msgConf)
		return
	}
	defer resp.Body.Close()

	// Читаем содержимое
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения файла: %v", err)
		msgConf := tgbotapi.NewMessage(int64(user.ChatID), "Ошибка чтения файла.")
		bot.Bot.Send(msgConf)
		return
	}

	// Импортируем активности
	err = db.ImportActivitiesFromYAML(buf.Bytes(), userID)
	if err != nil {
		log.Printf("Ошибка импорта активностей: %v", err)
		msgConf := tgbotapi.NewMessage(int64(user.ChatID),
			fmt.Sprintf("Ошибка импорта активностей: %v", err))
		bot.Bot.Send(msgConf)
		return
	}

	msgConf := tgbotapi.NewMessage(int64(user.ChatID),
		"✅ Активности успешно импортированы!")
	bot.Bot.Send(msgConf)
}

// DeleteActivityCommand обрабатывает команду удаления активности.
func DeleteActivityCommand(message *tgbotapi.Message) {
	tgUser := message.From
	if tgUser == nil {
		return
	}

	userID := common.UserID(tgUser.ID)

	user, err := db.GetUserByID(userID)
	if err != nil {
		log.Fatal(err)
	}

	msgText := "Выберите активность для удаления:"

	msgconf := tgbotapi.NewMessage(int64(user.ChatID), msgText)
	msgconf.ReplyMarkup = buildActivitiesKeyboardMarkupForUser(
		*user, -1, nil, nil, "delete_activity__delete", getDeleteActivitiesLastRow())

	_, err = bot.Bot.Send(msgconf)
	if err != nil {
		log.Fatal(err)
	}

	_, err = bot.Bot.Request(
		tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID),
	)
	if err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}
}

// DeleteActivityCallback обрабатывает callback для удаления активности.
func DeleteActivityCallback(callback *tgbotapi.CallbackQuery) {
	data := strings.Split(callback.Data, " ")
	if len(data) < 2 {
		return
	}

	activityID, err := strconv.ParseInt(data[1], 10, 64)
	if err != nil {
		log.Printf("Ошибка парсинга ID активности: %v", err)
		return
	}

	userID := common.UserID(callback.From.ID)

	// Получаем название активности перед удалением
	activityName, err := db.GetFullActivityNameByID(activityID, userID)
	if err != nil {
		log.Printf("Ошибка получения названия активности: %v", err)
		activityName = "неизвестная активность"
	}

	// Удаляем активность
	err = db.DeleteActivityRecursive(activityID, userID)
	if err != nil {
		log.Printf("Ошибка удаления активности: %v", err)

		answerConfig := tgbotapi.NewCallback(callback.ID, "Ошибка удаления активности")
		bot.Bot.Request(answerConfig)
		return
	}

	// Отправляем подтверждение
	msgText := fmt.Sprintf("✅ Активность '%s' и все её подактивности успешно удалены.", activityName)

	editConfig := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		msgText,
	)
	bot.Bot.Send(editConfig)

	answerConfig := tgbotapi.NewCallback(callback.ID, "Активность удалена")
	bot.Bot.Request(answerConfig)
}

// DeleteActivityCancelCallback отменяет удаление активности.
func DeleteActivityCancelCallback(callback *tgbotapi.CallbackQuery) {
	editConfig := tgbotapi.NewEditMessageText(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		"Удаление активности отменено.",
	)
	bot.Bot.Send(editConfig)

	answerConfig := tgbotapi.NewCallback(callback.ID, "Отменено")
	bot.Bot.Request(answerConfig)
}

// DeleteActivityRefreshCallback обновляет список активностей для удаления.
func DeleteActivityRefreshCallback(callback *tgbotapi.CallbackQuery) {
	userID := common.UserID(callback.From.ID)

	user, err := db.GetUserByID(userID)
	if err != nil {
		log.Printf("Ошибка получения пользователя: %v", err)
		return
	}

	msgText := "Выберите активность для удаления:"

	editConfig := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		msgText,
		buildActivitiesKeyboardMarkupForUser(
			*user, -1, nil, nil, "delete_activity__delete", getDeleteActivitiesLastRow()),
	)
	bot.Bot.Send(editConfig)

	answerConfig := tgbotapi.NewCallback(callback.ID, "Список обновлен")
	bot.Bot.Request(answerConfig)
}

// getDeleteActivitiesLastRow возвращает последний ряд кнопок для удаления активности.
func getDeleteActivitiesLastRow() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить", "delete_activity__refresh"),
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "delete_activity__cancel"),
	}
}
