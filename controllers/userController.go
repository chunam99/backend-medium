package controllers

import (
	config "backend/configs"
	"strconv"

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

func GetAllUsers(c *gin.Context) {
	type UserResponse struct {
		ID       uint   `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Photo    string `json:"photo"`
		Role     string `json:"role"`
	}

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "10")

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	perPageNum, err := strconv.Atoi(perPage)
	if err != nil || perPageNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid per_page number"})
		return
	}

	if perPageNum > 100 {
		perPageNum = 100
	}

	var totalUsers int64
	if err := config.DB.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}

	offset := (pageNum - 1) * perPageNum

	var users []models.User
	if err := config.DB.
		Limit(perPageNum).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	var response []UserResponse
	for _, user := range users {
		response = append(response, UserResponse{
			ID:       user.ID,
			Name:     user.Name,
			Username: user.Username,
			Photo:    user.Photo,
			Role:     user.Role,
		})
	}

	responses.PaginateResponse(c, response, totalUsers, pageNum, perPageNum)

}

func UpdateUserRole(c *gin.Context) {

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var request struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if request.Role != "admin" && request.Role != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role value"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Role = request.Role
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
		"user":    user,
	})
}
