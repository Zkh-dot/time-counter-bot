package db

import (
	"TimeCounterBot/common"
	"log"
	"time"
)

func AddActivityLog(activityLog ActivityLog) {
	database := getPostgreSQLDatabase()

	insertActivitySQL := `INSERT INTO activity_log (message_id, user_id, activity_id, timestamp, interval_minutes) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT(message_id, user_id)
		DO UPDATE SET activity_id = excluded.activity_id;
	`

	_, err := database.Exec(
		insertActivitySQL, activityLog.MessageID, activityLog.UserID,
		activityLog.ActivityID, activityLog.Timestamp, activityLog.IntervalMinutes,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func GetLogDurations(userID common.UserID, start, end time.Time) map[int64]float64 {
	// Запрос для суммирования длительностей по активности.
	const query = `
        SELECT activity_id, 
               COALESCE(SUM(interval_minutes), 0)
        FROM activity_log
        WHERE user_id = $1
          AND timestamp BETWEEN $2 AND $3
        GROUP BY activity_id
    `

	database := getPostgreSQLDatabase()
	rows, err := database.Query(query, userID, start, end)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Формируем мапу: activity_id -> суммарная длительность.
	logDurations := make(map[int64]float64)
	for rows.Next() {
		var actID int64
		var duration float64
		if err := rows.Scan(&actID, &duration); err != nil {
			log.Fatal(err)
		}
		logDurations[actID] = duration
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return logDurations
}
