package routes

import (
	tg "TimeCounterBot/tg/bot"
	"image/color"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
)

// Генерация кастомного графика
func generateAdvancedChart(filename string) error {
	p := plot.New()

	p.Title.Text = "Детальный график"
	p.X.Label.Text = "Время"
	p.Y.Label.Text = "Результаты"

	pts := plotter.XYs{
		{X: 1, Y: 10},
		{X: 2, Y: 20},
		{X: 3, Y: 15},
	}

	line, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}
	line.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Красная линия
	p.Add(line)

	return p.Save(400, 300, filename)
}

func GetDayStatisticsCommand(message *tgbotapi.Message) {
	filename := "advanced_chart.png"
	if err := generateAdvancedChart(filename); err != nil {
		log.Fatal("Ошибка генерации графика:", err)
	}

	msgconf := tgbotapi.NewPhoto(int64(message.Chat.ID), tgbotapi.FilePath(filename))
	_, err := tg.Bot.Send(msgconf)
	if err != nil {
		log.Fatal(err)
	}
}
