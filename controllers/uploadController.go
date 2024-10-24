package controllers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/storage" // Import for ACL handling
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type FirebaseStorage struct {
	App *firebase.App
}

// Initialize Firebase
func NewFirebaseStorage() (*FirebaseStorage, error) {
	ctx := context.Background()
	// Set the path to your Firebase service account key file
	sa := option.WithCredentialsFile("configs/serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase App: %v", err)
		return nil, err
	}
	log.Println("Firebase initialized successfully")

	return &FirebaseStorage{App: app}, nil
}

// UploadImage handles image uploading to Firebase Storage
func (fs *FirebaseStorage) UploadImage(c *gin.Context) {
	// Retrieve the file from the request
	file, header, err := c.Request.FormFile("upload")
	if err != nil {
		log.Println("Failed to retrieve file from request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not found in request"})
		return
	}
	defer file.Close()

	log.Printf("File received: %s", header.Filename)

	// Get the Firebase Storage client
	ctx := context.Background()
	client, err := fs.App.Storage(ctx)
	if err != nil {
		log.Println("Failed to get Firebase Storage client:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get storage client"})
		return
	}

	// Specify the bucket name
	bucketName := "blog-d2ef0.appspot.com"
	bucket, err := client.Bucket(bucketName)

	if err != nil {
		log.Println("Failed to get Specify the bucket name:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Specify the bucket name"})
		return
	}
	log.Printf("Using bucket: %s", bucketName)

	// Define the destination path in Firebase Storage
	objectName := fmt.Sprintf("images/%s", header.Filename)
	object := bucket.Object(objectName)

	log.Printf("Uploading file to Firebase Storage as: %s", objectName)

	// Create a new writer for the object
	wc := object.NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		log.Println("Failed to copy file to Firebase Storage writer:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	// Close the writer
	if err := wc.Close(); err != nil {
		log.Println("Failed to close Firebase Storage writer:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close writer"})
		return
	}

	log.Println("File uploaded successfully to Firebase Storage")

	// Set the file to be publicly readable
	if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		log.Println("Failed to set file to public:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set file to public"})
		return
	}

	log.Println("File set to public successfully")

	// Construct the public URL of the uploaded image
	imageURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)

	log.Printf("Public URL of uploaded image: %s", imageURL)

	// Respond with the URL
	c.JSON(http.StatusOK, gin.H{
		"uploaded": 1,
		"fileName": header.Filename,
		"url":      imageURL,
	})
}
