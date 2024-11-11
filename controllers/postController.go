package controllers

import (
	config "backend/configs"
	"backend/models"
	"backend/responses"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func GetAllPosts(c *gin.Context) {
	var posts []models.Post

	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "10")

	pageNum, _ := strconv.Atoi(page)
	perPageNum, _ := strconv.Atoi(perPage)

	var totalPosts int64
	config.DB.Model(&models.Post{}).Count(&totalPosts)

	offset := (pageNum - 1) * perPageNum

	if err := config.DB.Preload("User").Preload("Tags").Limit(perPageNum).Offset(offset).Find(&posts).Error; err != nil {
		c.JSON(500, gin.H{"error": "Could not retrieve posts"})
		return
	}

	var postResponses []map[string]interface{}
	for _, post := range posts {
		userResponse := models.ToUserResponse(post.User)

		postResponse := map[string]interface{}{
			"id":          post.ID,
			"title":       post.Title,
			"description": post.Description,
			"content":     post.Content,
			"image":       post.Image,
			"pinned":      post.Pinned,
			"claps":       post.Claps,
			"tags":        post.Tags,
			"comment":     post.Comment,
			"created_at":  post.CreatedAt,
			"updated_at":  post.UpdatedAt,
			"user":        userResponse,
		}

		postResponses = append(postResponses, postResponse)
	}

	responses.PaginateResponse(c, postResponses, totalPosts, pageNum, perPageNum)
}
func GetPinnedPosts(c *gin.Context) {
	var posts []models.Post

	if err := config.DB.Preload("User").Where("pinned = ?", true).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve pinned posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pinnedPosts": posts})
}

func GetPostByID(c *gin.Context) {
	postID := c.Param("id")
	var post models.Post

	if err := config.DB.Preload("User").Preload("Tags").Where("id = ?", postID).First(&post).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	userResponse := models.ToUserResponse(post.User)

	postResponse := map[string]interface{}{
		"id":          post.ID,
		"title":       post.Title,
		"description": post.Description,
		"content":     post.Content,
		"image":       post.Image,
		"pinned":      post.Pinned,
		"claps":       post.Claps,
		"tags":        post.Tags,
		"comment":     post.Comment,
		"created_at":  post.CreatedAt,
		"updated_at":  post.UpdatedAt,
		"user":        userResponse,
	}

	c.JSON(http.StatusOK, gin.H{"data": postResponse})

}

func CreatePost(c *gin.Context) {
	var request models.CreatePostRequest

	if err := c.ShouldBindWith(&request, binding.FormMultipart); err != nil {
		log.Println("Error parsing form data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data", "details": err.Error()})
		return
	}

	image, err := c.FormFile("image")
	if err == nil {
		log.Println("Received image:", image.Filename)
	}

	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID := userIDInterface.(uint)

	var tags []models.Tag
	for _, tagName := range request.Tags {
		var tag models.Tag

		if err := config.DB.Where("name = ?", tagName).First(&tag).Error; err != nil {
			tag = models.Tag{Name: tagName}
			if err := config.DB.Create(&tag).Error; err != nil {
				log.Println("Error creating tag:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create tag"})
				return
			}
		}
		tags = append(tags, tag)
	}

	post := models.Post{
		Title:       request.Title,
		Description: request.Description,
		Content:     request.Content,
		UserID:      userID,
		Tags:        tags,
	}

	log.Println("userData.ID", userID)

	if err := config.DB.Create(&post).Error; err != nil {
		log.Println("Error saving post:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save the post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post created successfully", "post": post})
}

func UpdatePost(c *gin.Context) {
	postID := c.Param("id")
	var post models.Post

	if err := config.DB.Where("id = ?", postID).First(&post).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists || post.UserID != userID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to update this post"})
		return
	}

	var updatedPost models.Post
	if err := c.ShouldBindJSON(&updatedPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.Title = updatedPost.Title
	post.Content = updatedPost.Content
	post.Image = updatedPost.Image

	if err := config.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post updated successfully", "post": post})
}

func DeletePost(c *gin.Context) {
	postID := c.Param("id")
	var post models.Post

	if err := config.DB.Where("id = ?", postID).First(&post).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists || post.UserID != userID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission to delete this post"})
		return
	}

	if err := config.DB.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

func ClapPost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var requestBody struct {
		Claps int `json:"claps"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var post models.Post
	if err := config.DB.First(&post, uint(postID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	post.Claps += uint(requestBody.Claps)

	if err := config.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update claps"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message": "Clapped successfully",
			"claps":   post.Claps,
		},
	})
}

func GetPostClaps(c *gin.Context) {
	postID := c.Param("id")
	var post models.Post

	if err := config.DB.Where("id = ?", postID).First(&post).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"totalClaps": post.Claps})
}
