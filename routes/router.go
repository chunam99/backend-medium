package routes

import (
	"backend/controllers"
	"backend/middleware"
	"log"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	log.Println("Setting up routes...")

	firebaseStorage, err := controllers.NewFirebaseStorage()
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Storage: %v", err)
	}

	// Define the upload route
	r.POST("/upload", firebaseStorage.UploadImage)
	r.POST("/register", controllers.RegisterUser)
	r.POST("/login", controllers.LoginUser)

	r.GET("/posts", controllers.GetAllPosts)
	r.GET("/posts/pinned", controllers.GetPinnedPosts)
	r.GET("/posts/:id", controllers.GetPostByID)

	authorized := r.Group("/")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.POST("/posts", controllers.CreatePost)
		authorized.PUT("/posts/:id", controllers.UpdatePost)
		authorized.DELETE("/posts/:id", controllers.DeletePost)
	}
	log.Println("Routes setup complete.")

}
