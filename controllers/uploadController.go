package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	"google.golang.org/api/option"
)

type FirebaseStorage struct {
	App *firebase.App
}

// Initialize Firebase
func NewFirebaseStorage() (*FirebaseStorage, error) {
	ctx := context.Background()

	serviceAccountBase64 := os.Getenv("SERVICE_ACCOUNT_KEY")
	if serviceAccountBase64 == "" {
		log.Fatal("SERVICE_ACCOUNT_KEY is not set")
		return nil, fmt.Errorf("service account key not set")
	}

	decodedKey, err := base64.StdEncoding.DecodeString(serviceAccountBase64)
	if err != nil {
		log.Fatalf("Failed to decode service account key: %v", err)
		return nil, err
	}

	tempFilePath := "configs/serviceAccountKey.json"
	err = os.WriteFile(tempFilePath, decodedKey, 0644)
	if err != nil {
		log.Fatalf("Failed to write service account key file: %v", err)
		return nil, err
	}

	sa := option.WithCredentialsFile(tempFilePath)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase App: %v", err)
		return nil, err
	}
	log.Println("Firebase initialized successfully")

	return &FirebaseStorage{App: app}, nil
}

// ResizeImage resizes an image to a given width, maintaining aspect ratio
func ResizeImage(file io.Reader, width uint) (io.Reader, error) {
	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Resize the image
	resizedImg := resize.Resize(width, 0, img, resize.Lanczos3)

	// Encode the resized image into a buffer
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, resizedImg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %v", err)
	}

	return buf, nil
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

	// Resize the image before uploading
	resizedFile, err := ResizeImage(file, 800) // Resize to 800px width
	if err != nil {
		log.Println("Failed to resize image:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resize image"})
		return
	}

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
		log.Println("Failed to get bucket:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bucket"})
		return
	}
	log.Printf("Using bucket: %s", bucketName)

	// Define the destination path in Firebase Storage
	objectName := fmt.Sprintf("images/%s", header.Filename)
	object := bucket.Object(objectName)

	log.Printf("Uploading file to Firebase Storage as: %s", objectName)

	// Create a new writer for the object
	wc := object.NewWriter(ctx)
	if _, err := io.Copy(wc, resizedFile); err != nil {
		log.Println("Failed to copy resized image to Firebase Storage writer:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload resized image"})
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
