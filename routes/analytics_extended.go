package routes

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"TimeCounterBot/common"
	"TimeCounterBot/db"
	"TimeCounterBot/tg/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AnalyticsMenuCommand показывает главное меню аналитики.
func AnalyticsMenuCommand(message *tgbotapi.Message) {
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

	msgText := "📊 *Аналитика активностей*\n\nВыберите тип отчета:"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 Статистика за период", "analytics__day_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Сравнить периоды", "analytics__compare_periods"),
		),
	)

	msgConf := tgbotapi.NewMessage(int64(user.ChatID), msgText)
	msgConf.ParseMode = "Markdown"
	msgConf.ReplyMarkup = keyboard

	_, err = bot.Bot.Send(msgConf)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}

	// Удаляем исходное сообщение команды
	_, err = bot.Bot.Request(
		tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID),
	)
	if err != nil {
		log.Printf("Ошибка удаления сообщения: %v", err)
	}
}

// AnalyticsGetDayStatsCallback перенаправляет на существующую функцию статистики.
func AnalyticsGetDayStatsCallback(callback *tgbotapi.CallbackQuery) {
	// Создаем фиктивное сообщение для передачи в GetDayStatisticsCommand
	message := &tgbotapi.Message{
		From: callback.From,
		Chat: callback.Message.Chat,
	}

	GetDayStatisticsCommand(message)

	answerConfig := tgbotapi.NewCallback(callback.ID, "Открываем статистику...")
	bot.Bot.Request(answerConfig)
}

// AnalyticsComperiodsCallback показывает меню выбора периодов для сравнения.
func AnalyticsComperiodsCallback(callback *tgbotapi.CallbackQuery) {
	msgText := "📊 *Сравнение периодов*\n\nВыберите, какие периоды хотите сравнить:"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Эта неделя vs прошлая", "compare_periods__this_vs_last_week"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📆 Этот месяц vs прошлый", "compare_periods__this_vs_last_month"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔧 Настроить периоды", "compare_periods__custom"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "analytics__back"),
		),
	)

	editConfig := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		msgText,
		keyboard,
	)
	editConfig.ParseMode = "Markdown"
	bot.Bot.Send(editConfig)

	answerConfig := tgbotapi.NewCallback(callback.ID, "Выберите период для сравнения")
	bot.Bot.Request(answerConfig)
}

// AnalyticsBackCallback возвращает к главному меню аналитики.
func AnalyticsBackCallback(callback *tgbotapi.CallbackQuery) {
	msgText := "📊 *Аналитика активностей*\n\nВыберите тип отчета:"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 Статистика за период", "analytics__day_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Сравнить периоды", "analytics__compare_periods"),
		),
	)

	editConfig := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		msgText,
		keyboard,
	)
	editConfig.ParseMode = "Markdown"
	bot.Bot.Send(editConfig)

	answerConfig := tgbotapi.NewCallback(callback.ID, "Главное меню аналитики")
	bot.Bot.Request(answerConfig)
}

// ComparePeriods_ThisVsLastWeekCallback сравнивает текущую и прошлую неделю.
func ComparePeriods_ThisVsLastWeekCallback(callback *tgbotapi.CallbackQuery) {
	userID := common.UserID(callback.From.ID)

	now := time.Now()

	// Текущая неделя (понедельник - воскресенье)
	weekday := int(now.Weekday())
	if weekday == 0 { // Воскресенье
		weekday = 7
	}
	thisWeekStart := now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
	thisWeekEnd := thisWeekStart.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// Прошлая неделя
	lastWeekStart := thisWeekStart.AddDate(0, 0, -7)
	lastWeekEnd := thisWeekEnd.AddDate(0, 0, -7)

	comparison, err := db.CompareActivityPeriods(
		userID,
		thisWeekStart, thisWeekEnd,
		lastWeekStart, lastWeekEnd,
		"Эта неделя",
		"Прошлая неделя",
	)

	if err != nil {
		log.Printf("Ошибка сравнения периодов: %v", err)
		answerConfig := tgbotapi.NewCallback(callback.ID, "Ошибка получения данных")
		bot.Bot.Request(answerConfig)
		return
	}

	msgText := formatComparisonResult(comparison)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "compare_periods__back"),
		),
	)

	editConfig := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		msgText,
		keyboard,
	)
	editConfig.ParseMode = "Markdown"
	bot.Bot.Send(editConfig)

	answerConfig := tgbotapi.NewCallback(callback.ID, "Сравнение выполнено")
	bot.Bot.Request(answerConfig)
}

// ComparePeriods_ThisVsLastMonthCallback сравнивает текущий и прошлый месяц.
func ComparePeriods_ThisVsLastMonthCallback(callback *tgbotapi.CallbackQuery) {
	userID := common.UserID(callback.From.ID)

	now := time.Now()

	// Текущий месяц
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	thisMonthEnd := thisMonthStart.AddDate(0, 1, 0).Add(-time.Second)

	// Прошлый месяц
	lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
	lastMonthEnd := thisMonthStart.Add(-time.Second)

	comparison, err := db.CompareActivityPeriods(
		userID,
		thisMonthStart, thisMonthEnd,
		lastMonthStart, lastMonthEnd,
		"Этот месяц",
		"Прошлый месяц",
	)

	if err != nil {
		log.Printf("Ошибка сравнения периодов: %v", err)
		answerConfig := tgbotapi.NewCallback(callback.ID, "Ошибка получения данных")
		bot.Bot.Request(answerConfig)
		return
	}

	msgText := formatComparisonResult(comparison)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "compare_periods__back"),
		),
	)

	editConfig := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		msgText,
		keyboard,
	)
	editConfig.ParseMode = "Markdown"
	bot.Bot.Send(editConfig)

	answerConfig := tgbotapi.NewCallback(callback.ID, "Сравнение выполнено")
	bot.Bot.Request(answerConfig)
}

// ComparePeriods_CustomCallback показывает инструкции для настройки периодов.
func ComparePeriods_CustomCallback(callback *tgbotapi.CallbackQuery) {
	msgText := "🔧 *Настраиваемое сравнение*\n\n" +
		"Эта функция пока не реализована.\n" +
		"В будущем здесь можно будет выбрать произвольные даты для сравнения."

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "compare_periods__back"),
		),
	)

	editConfig := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		msgText,
		keyboard,
	)
	editConfig.ParseMode = "Markdown"
	bot.Bot.Send(editConfig)

	answerConfig := tgbotapi.NewCallback(callback.ID, "Функция в разработке")
	bot.Bot.Request(answerConfig)
}

// ComparePeriods_BackCallback возвращает к меню сравнения периодов.
func ComparePeriods_BackCallback(callback *tgbotapi.CallbackQuery) {
	msgText := "📊 *Сравнение периодов*\n\nВыберите, какие периоды хотите сравнить:"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Эта неделя vs прошлая", "compare_periods__this_vs_last_week"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📆 Этот месяц vs прошлый", "compare_periods__this_vs_last_month"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔧 Настроить периоды", "compare_periods__custom"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "analytics__back"),
		),
	)

	editConfig := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		msgText,
		keyboard,
	)
	editConfig.ParseMode = "Markdown"
	bot.Bot.Send(editConfig)

	answerConfig := tgbotapi.NewCallback(callback.ID, "Меню сравнения периодов")
	bot.Bot.Request(answerConfig)
}

// formatComparisonResult форматирует результат сравнения в красивый текст.
func formatComparisonResult(comparison *db.PeriodComparisonResult) string {
	if len(comparison.Comparisons) == 0 {
		return "📊 *Сравнение периодов*\n\nНет данных для сравнения."
	}

	// Сортируем по убыванию разности во времени
	sort.Slice(comparison.Comparisons, func(i, j int) bool {
		return math.Abs(float64(comparison.Comparisons[i].DifferenceMin)) > math.Abs(float64(comparison.Comparisons[j].DifferenceMin))
	})

	result := fmt.Sprintf("📊 *Сравнение: %s vs %s*\n\n",
		comparison.Period1Name, comparison.Period2Name)

	// Общая статистика
	totalDiff := comparison.Period1Total - comparison.Period2Total
	totalDiffHours := float64(totalDiff) / 60.0
	result += fmt.Sprintf("⏱ *Общее время:*\n")
	result += fmt.Sprintf("• %s: %s\n", comparison.Period1Name, formatMinutes(comparison.Period1Total))
	result += fmt.Sprintf("• %s: %s\n", comparison.Period2Name, formatMinutes(comparison.Period2Total))

	if totalDiff > 0 {
		result += fmt.Sprintf("📈 *Изменение:* +%.1f ч\n\n", totalDiffHours)
	} else if totalDiff < 0 {
		result += fmt.Sprintf("📉 *Изменение:* %.1f ч\n\n", totalDiffHours)
	} else {
		result += fmt.Sprintf("➖ *Изменение:* без изменений\n\n")
	}

	// Топ изменений (максимум 5 активностей)
	result += "*Основные изменения:*\n"

	count := 0
	for _, comp := range comparison.Comparisons {
		if count >= 5 {
			break
		}

		if comp.DifferenceMin == 0 {
			continue // Пропускаем активности без изменений
		}

		icon := "📈"
		sign := "+"
		if comp.DifferenceMin < 0 {
			icon = "📉"
			sign = ""
		} else if comp.Period2Minutes == 0 {
			icon = "🆕"
		} else if comp.Period1Minutes == 0 {
			icon = "❌"
		}

		diffHours := float64(comp.DifferenceMin) / 60.0
		percentStr := ""
		if comp.PercentChange != 0 {
			if comp.PercentChange == 100 {
				percentStr = " (новая)"
			} else if comp.PercentChange == -100 {
				percentStr = " (исчезла)"
			} else {
				percentStr = fmt.Sprintf(" (%s%.0f%%)", sign, math.Abs(comp.PercentChange))
			}
		}

		result += fmt.Sprintf("%s *%s*: %s%.1f ч%s\n",
			icon, comp.ActivityName, sign, math.Abs(diffHours), percentStr)

		count++
	}

	if count == 0 {
		result += "Значительных изменений не обнаружено."
	}

	return result
}

// formatMinutes форматирует минуты в часы и минуты.
func formatMinutes(minutes int64) string {
	if minutes == 0 {
		return "0 мин"
	}

	hours := minutes / 60
	mins := minutes % 60

	if hours == 0 {
		return fmt.Sprintf("%d мин", mins)
	} else if mins == 0 {
		return fmt.Sprintf("%d ч", hours)
	} else {
		return fmt.Sprintf("%d ч %d мин", hours, mins)
	}
}
