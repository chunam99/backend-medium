package controllers

import (
	config "backend/configs"

	"backend/models"
	"backend/responses"
	"backend/services"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("cvn.dev.jwt.key.2024")

type Claims struct {
	UserID   uint   `json:"sub"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Photo    string `json:"avatar"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func RegisterUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	if strings.TrimSpace(user.Username) == "" {
		log.Println("Username validation failed")
		responses.ErrorResponse(c, http.StatusBadRequest, "Username cannot be empty", "Validation error")
		return
	}

	if len(user.Password) < 8 {
		log.Println("Password validation failed")
		responses.ErrorResponse(c, http.StatusBadRequest, "Password must be at least 8 characters long", "Validation error")
		return
	}

	var existingUser models.User
	result := config.DB.Where("username = ?", user.Username).First(&existingUser)
	if result.RowsAffected > 0 {
		log.Println("Username already exists")
		responses.ErrorResponse(c, http.StatusBadRequest, "Username already exists", "Username is already taken")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		log.Println("Failed to hash password")
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}
	user.Password = string(hashedPassword)
	user.Role = "user"

	log.Printf("User data: %+v", user)
	if err := config.DB.Create(&user).Error; err != nil {
		log.Printf("Error saving user: %v", err)
		responses.ErrorResponse(c, http.StatusInternalServerError, "Could not save user", err.Error())
		return
	}

	responses.SuccessResponse(c, "Registration successful", user)
}

func LoginUser(c *gin.Context) {
	var request struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		responses.ErrorResponse(c, http.StatusBadRequest, "Invalid request", "Invalid input data")
		return
	}

	user, err := services.AuthenticateUser(request.Username, request.Password)
	if err != nil || user == nil {
		responses.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials", "Authentication failed")
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Subject:   fmt.Sprintf("%d", user.ID), // Nên để đây là ID của user
		},
	}
	claims.Name = user.Name
	claims.Photo = user.Photo
	claims.Role = user.Role

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		responses.ErrorResponse(c, http.StatusInternalServerError, "Failed to create token", err.Error())
		return
	}

	responses.SuccessResponse(c, "Login successful", gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"username": user.Username,
			"photo":    user.Photo,
		},
	})
}
