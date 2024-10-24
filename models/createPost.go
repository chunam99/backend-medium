package models

type CreatePostRequest struct {
	Title       string   `form:"title" binding:"required"`
	Description string   `form:"desc" binding:"required"`
	Content     string   `form:"content" binding:"required"`
	Image       *string  `form:"image"`
	Categories  []string `form:"categories,omitempty"`
	Tags        []string `form:"tags,omitempty"`
}
