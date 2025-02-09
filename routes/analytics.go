package routes

import (
	tg "TimeCounterBot/tg/bot"
	"encoding/json"
	"log"
	"os/exec"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Node представляет узел дерева активностей
type Node struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	ParentID *int     `json:"parent_id,omitempty"`
	Duration *float64 `json:"duration,omitempty"`
}

// ChartData содержит список узлов для сериализации
type ChartData struct {
	Nodes []Node `json:"nodes"`
}

// GetDayStatisticsCommand вызывается, когда пользователь запрашивает статистику
func GetDayStatisticsCommand(message *tgbotapi.Message) {
	// Дерево активностей
	data := ChartData{
		Nodes: []Node{
			{ID: 1, Name: "Work", ParentID: nil, Duration: nil},
			{ID: 2, Name: "Sleep", ParentID: nil, Duration: floatPtr(6)},
			{ID: 3, Name: "Leisure", ParentID: nil, Duration: nil},
			{ID: 4, Name: "Exercise", ParentID: nil, Duration: floatPtr(2)},
			{ID: 5, Name: "Breakfast", ParentID: intPtr(1), Duration: floatPtr(2)},
			{ID: 6, Name: "Programming", ParentID: intPtr(1), Duration: nil},
			{ID: 7, Name: "Golang", ParentID: intPtr(6), Duration: floatPtr(4)},
			{ID: 8, Name: "Python", ParentID: intPtr(6), Duration: floatPtr(2)},
			{ID: 9, Name: "Gaming", ParentID: intPtr(3), Duration: floatPtr(5)},
		},
	}

	// Кодируем данные в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Ошибка кодирования JSON: %v", err)
	}

	// Путь к Python-скрипту
	scriptPath := "python_scripts/generate_sunburst_chart.py"
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

// Вспомогательные функции для указателей
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
