package db

import (
	"errors"
	"log"
	"strconv"

	"TimeCounterBot/common"
)

func AddUser(user User) {
	// mutex.Lock()
	// defer mutex.Unlock()

	// database, err := sql.Open("sqlite3", "database.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer database.Close()

	database := getPostgreSQLDatabase()

	insertUserSQL := `INSERT INTO users 
			(id, chat_id, timer_enabled, timer_minutes, 
			schedule_morning_start_hour, schedule_evening_finish_hour, last_notify) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := database.Exec(
		insertUserSQL, user.ID, user.ChatID, user.TimerEnabled, user.TimerMinutes,
		user.ScheduleMorningStartHour, user.ScheduleEveningFinishHour, user.LastNotify,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func GetUserByID(userID common.UserID) (*User, error) {
	// mutex.Lock()
	// defer mutex.Unlock()

	// database, err := sql.Open("sqlite3", "database.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer database.Close()

	database := getPostgreSQLDatabase()

	selectUserSQL := `SELECT id, chat_id, timer_enabled, timer_minutes, 
			schedule_morning_start_hour, schedule_evening_finish_hour, last_notify FROM users
		WHERE id = $1
	`

	rows, err := database.Query(selectUserSQL, userID)
	if err != nil {
		log.Fatal(err)
	}
	if rows.Err() != nil {
		log.Fatal(rows.Err())
	}
	defer rows.Close()

	found := false

	user := User{}
	for rows.Next() {
		err = rows.Scan(
			&user.ID, &user.ChatID, &user.TimerEnabled, &user.TimerMinutes,
			&user.ScheduleMorningStartHour, &user.ScheduleEveningFinishHour, &user.LastNotify,
		)
		if err != nil {
			log.Fatal(err)
		}

		found = true
	}

	if !found {
		return nil, errors.New("User with id " + strconv.FormatInt(int64(userID), 10) + " was not found.")
	}

	return &user, nil
}

func UpdateUser(user User) {
	// mutex.Lock()
	// defer mutex.Unlock()

	// database, err := sql.Open("sqlite3", "database.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer database.Close()

	database := getPostgreSQLDatabase()

	updateUserSQL := `UPDATE users
		SET timer_enabled = $1, timer_minutes = $2, 
		  schedule_morning_start_hour = $3, schedule_evening_finish_hour = $4,
		  last_notify = $5
		WHERE id = $6
	`

	_, err := database.Exec(
		updateUserSQL, user.TimerEnabled, user.TimerMinutes, user.ScheduleMorningStartHour,
		user.ScheduleEveningFinishHour, user.LastNotify, user.ID,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func GetUsers() []User {
	// mutex.Lock()
	// defer mutex.Unlock()

	// database, err := sql.Open("sqlite3", "database.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer database.Close()

	database := getPostgreSQLDatabase()

	selectUserSQL := `SELECT id, chat_id, timer_enabled, timer_minutes, 
		schedule_morning_start_hour, schedule_evening_finish_hour, last_notify 
	FROM users`

	rows, err := database.Query(selectUserSQL)
	if err != nil {
		log.Fatal(err)
	}
	if rows.Err() != nil {
		log.Fatal(rows.Err())
	}
	defer rows.Close()

	users := make([]User, 0)

	for rows.Next() {
		user := User{}

		err = rows.Scan(
			&user.ID, &user.ChatID, &user.TimerEnabled, &user.TimerMinutes,
			&user.ScheduleMorningStartHour, &user.ScheduleEveningFinishHour, &user.LastNotify,
		)
		if err != nil {
			log.Fatal(err)
		}

		users = append(users, user)
	}

	return users
}
