package routes

import (
	"log"
	"time"

	"TimeCounterBot/db"
)

const DispatchInterval = time.Second * 5

func isTimeInInterval(ts time.Time, startHour, finishHour int64) bool {
	return ts.Hour() >= int(startHour) && ts.Hour() < int(finishHour)
}

func DispatchNotifications() {
	now := time.Now()

	users, err := db.GetUsers()
	if err != nil {
		log.Fatal(err)
	}
	for _, user := range users {
		if !user.TimerEnabled {
			continue
		}

		if !user.ScheduleMorningStartHour.Valid || !user.ScheduleEveningFinishHour.Valid || !user.TimerMinutes.Valid {
			log.Fatalf("something is invalid for user %d", user.ID)
		}

		startHour := user.ScheduleMorningStartHour.Int64
		finishHour := user.ScheduleEveningFinishHour.Int64
		if !isTimeInInterval(now, startHour, finishHour) {
			continue
		}

		if user.LastNotify.Valid && now.Sub(user.LastNotify.Time) < time.Minute*time.Duration(user.TimerMinutes.Int64) {
			continue
		}

		go notifyUser(user)
		if !isTimeInInterval(now.Add(time.Minute*time.Duration(user.TimerMinutes.Int64)), startHour, finishHour) {

		}
	}

	time.Sleep(DispatchInterval)

	go DispatchNotifications()
}
