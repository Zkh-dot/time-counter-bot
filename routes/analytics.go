package routes

import (
	tg "TimeCounterBot/tg/bot"
	"encoding/json"
	"log"
	"os/exec"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GetDayStatisticsCommand вызывается, когда пользователь запрашивает статистику
func GetDayStatisticsCommand(message *tgbotapi.Message) {
	// Данные для диаграммы (пример)
	data := map[string]float64{
		"Work":     8,
		"Sleep":    6,
		"Leisure":  5,
		"Exercise": 2,
		"Eating":   3,
	}

	// Кодируем данные в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Ошибка кодирования JSON: %v", err)
	}

	// Путь к Python-скрипту
	scriptName := "generate_pie_chart.py"
	scriptPath := "python_scripts/" + scriptName
	outputFile := "pie_chart.png"

	// Запускаем Python-скрипт
	cmd := exec.Command("python3", scriptPath, string(jsonData), outputFile)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Ошибка выполнения скрипта: %v", err)
	}

	// Отправляем картинку в Telegram
	msgconf := tgbotapi.NewPhoto(int64(message.Chat.ID), tgbotapi.FilePath(outputFile))
	_, err = tg.Bot.Send(msgconf)
	if err != nil {
		log.Fatalf("Ошибка отправки изображения: %v", err)
	}
}
