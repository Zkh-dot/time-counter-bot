package db

import (
	"TimeCounterBot/common"
)

// AddUser добавляет нового пользователя в базу.
func AddUser(user User) error {
	result := GormDB.Create(&user)
	return result.Error
}

// GetUserByID возвращает пользователя по его идентификатору.
func GetUserByID(userID common.UserID) (*User, error) {
	var user User
	result := GormDB.First(&user, "id = ?", userID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// UpdateUser обновляет данные пользователя.
func UpdateUser(user User) error {
	result := GormDB.Save(&user)
	return result.Error
}

// GetUsers возвращает список всех пользователей.
func GetUsers() ([]User, error) {
	var users []User
	result := GormDB.Find(&users)
	return users, result.Error
}
