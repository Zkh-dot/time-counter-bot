package routes

import (
	tg "TimeCounterBot/tg/bot"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

// Data for the pie chart
var data = map[string]float64{
	"Work":     8,
	"Sleep":    6,
	"Leisure":  5,
	"Exercise": 2,
	"Eating":   3,
}

// Colors for each section
var colors = []color.Color{
	color.RGBA{255, 99, 132, 255},  // Red
	color.RGBA{54, 162, 235, 255},  // Blue
	color.RGBA{255, 206, 86, 255},  // Yellow
	color.RGBA{75, 192, 192, 255},  // Green
	color.RGBA{153, 102, 255, 255}, // Purple
}

// Function to draw a pie slice
func drawPieSlice(dc draw.Canvas, center vg.Point, radius vg.Length, startAngle, endAngle float64, col color.Color) {
	path := vg.Path{}
	path.Move(vg.Point{center.X, center.Y})

	// Создаем точки по кругу
	for angle := startAngle; angle <= endAngle; angle += 0.01 {
		x := center.X + radius*vg.Length(math.Cos(angle))
		y := center.Y + radius*vg.Length(math.Sin(angle))
		path.Line(vg.Point{x, y})
	}

	path.Close()
	dc.SetColor(col)
	dc.Fill(path)
}

func generateAdvancedChart(filename string) {
	// Создаём новый график
	p := plot.New()
	p.HideAxes() // Убираем оси

	// Вычисляем сумму всех значений
	total := 0.0
	for _, v := range data {
		total += v
	}

	// Начальный угол
	startAngle := 0.0
	center := vg.Point{X: 4 * vg.Inch, Y: 4 * vg.Inch}
	radius := 3 * vg.Inch

	// Создаем canvas и рисуем на нём
	c := vgimg.New(8*vg.Inch, 8*vg.Inch)
	dc := draw.NewCanvas(c, 8*vg.Inch, 8*vg.Inch)

	// Рисуем сектора
	i := 0
	for _, value := range data {
		angle := (value / total) * 2 * math.Pi
		drawPieSlice(dc, center, radius, startAngle, startAngle+angle, colors[i%len(colors)])
		startAngle += angle
		i++
	}

	// Открываем файл для сохранения
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Ошибка создания файла:", err)
		return
	}
	defer f.Close()

	// Кодируем изображение в PNG и записываем в файл
	if err := png.Encode(f, c.Image()); err != nil {
		fmt.Println("Ошибка при сохранении PNG:", err)
	} else {
		fmt.Println("Круговая диаграмма сохранена:", filename)
	}
}

func GetDayStatisticsCommand(message *tgbotapi.Message) {
	filename := "advanced_chart.png"
	generateAdvancedChart(filename)

	msgconf := tgbotapi.NewPhoto(int64(message.Chat.ID), tgbotapi.FilePath(filename))
	_, err := tg.Bot.Send(msgconf)
	if err != nil {
		log.Fatal(err)
	}
}
