package routes

import (
	"log"
	"time"

	"TimeCounterBot/db"
)

func DispatchNotifications() {
	now := time.Now()

	users := db.GetUsers()
	for _, user := range users {
		if !user.TimerEnabled {
			continue
		}

		if !user.ScheduleMorningStartHour.Valid || !user.ScheduleEveningFinishHour.Valid || !user.TimerMinutes.Valid {
			log.Fatalf("something is invalid for user %d", user.ID)
		}

		if now.Hour() < int(user.ScheduleMorningStartHour.Int64) || now.Hour() >= int(user.ScheduleEveningFinishHour.Int64) {
			continue
		}

		if user.LastNotify.Valid && now.Sub(user.LastNotify.Time) < time.Minute*time.Duration(user.TimerMinutes.Int64) {
			continue
		}

		go notifyUser(user)
	}

	time.Sleep(time.Second * 5)

	go DispatchNotifications()
}
