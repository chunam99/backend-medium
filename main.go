package main

import (
	config "backend/configs"
	"backend/models"
	"backend/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	config.ConnectDatabase()

	config.DB.AutoMigrate(&models.User{}, &models.Post{}, &models.Clap{})

	r := gin.Default()

	r.Use(config.SetupCORS())

	routes.SetupRouter(r)
	log.Println("Attempting to start server on :8386...")
	err = r.Run(":8386")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	} else {
		log.Println("Server started on :8386")
	}

}
