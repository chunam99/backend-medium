package services

import (
	config "backend/configs"

	"backend/models"

	"golang.org/x/crypto/bcrypt"
)

func AuthenticateUser(username, password string) (*models.User, error) {
	var user models.User

	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	return &user, nil
}
