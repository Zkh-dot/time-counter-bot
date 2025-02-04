package common

type (
	UserID int64
	ChatID int64
)

type UserStateSimple int

type UserState struct {
	State          UserStateSimple
	WaitingChannel *chan string
}

const (
	Idle UserStateSimple = iota
	InCommand
)

var UserStates = make(map[UserID]UserState)
