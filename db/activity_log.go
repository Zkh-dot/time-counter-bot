package db

import (
	"log"
)

func AddActivityLog(activityLog ActivityLog) {
	// mutex.Lock()
	// defer mutex.Unlock()

	// database, err := sql.Open("sqlite3", "database.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer database.Close()

	database := getPostgreSQLDatabase()

	insertActivitySQL := `INSERT INTO activity_log (message_id, user_id, activity_id, timestamp, interval_minutes) 
		VALUES (?, ?, ?, ?, ?)
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
