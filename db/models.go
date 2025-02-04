package db

import (
	"database/sql"
	"sync"
	"time"

	"TimeCounterBot/common"
)

var mutex sync.Mutex

type Activity struct {
	ID               int64
	UserID           int64
	Name             string
	ParentActivityID int64
	IsLeaf           bool
}

type ActivityLog struct {
	MessageID       int64
	UserID          int64
	ActivityID      int64
	Timestamp       time.Time
	IntervalMinutes int64
}

type User struct {
	ID                        common.UserID
	ChatID                    common.ChatID
	TimerEnabled              bool
	TimerMinutes              sql.NullInt64
	ScheduleMorningStartHour  sql.NullInt64
	ScheduleEveningFinishHour sql.NullInt64
	LastNotify                sql.NullTime
}

type ActivityRoute struct {
	Name   string
	LeafID int64
}
