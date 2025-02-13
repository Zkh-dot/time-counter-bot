package db

import (
	"time"

	"TimeCounterBot/common"

	"gorm.io/gorm/clause"
)

// AddActivityLog добавляет лог активности. При конфликте по (message_id, user_id)
// обновляет поле activity_id.
func AddActivityLog(activityLog ActivityLog) error {
	result := GormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "message_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"activity_id"}),
	}).Create(&activityLog)
	return result.Error
}

// GetLogDurations получает суммарную длительность для каждой активности
// для пользователя userID за интервал [start, end].
func GetLogDurations(userID common.UserID, start, end time.Time) (map[int64]float64, error) {
	var results []struct {
		ActivityID    int64
		TotalInterval int64
	}

	err := GormDB.Model(&ActivityLog{}).
		Select("activity_id, COALESCE(SUM(interval_minutes), 0) as total_interval").
		Where("user_id = ? AND timestamp BETWEEN ? AND ?", userID, start, end).
		Group("activity_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	logDurations := make(map[int64]float64)
	for _, r := range results {
		logDurations[r.ActivityID] = float64(r.TotalInterval)
	}
	return logDurations, nil
}
