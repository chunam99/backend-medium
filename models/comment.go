package models

import "time"

type Comment struct {
	ID          uint   `gorm:"primaryKey"`
	CommentText string `gorm:"not null"`
	UserID      uint   `gorm:"not null"`
	PostID      uint   `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
