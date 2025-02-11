package db

import (
	"TimeCounterBot/common"
)

// AddUser добавляет нового пользователя в базу.
func AddUser(user User) error {
	result := DB.Create(&user)
	return result.Error
}

// GetUserByID возвращает пользователя по его идентификатору.
func GetUserByID(userID common.UserID) (*User, error) {
	var user User
	result := DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// UpdateUser обновляет данные пользователя.
func UpdateUser(user User) error {
	result := DB.Save(&user)
	return result.Error
}

// GetUsers возвращает список всех пользователей.
func GetUsers() ([]User, error) {
	var users []User
	result := DB.Find(&users)
	return users, result.Error
}
