package models

import "time"

type Tag struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

type Post struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `gorm:"not null" json:"description"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Image       string    `gorm:"size:255" json:"image"`
	Pinned      bool      `gorm:"default:false" json:"pinned"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user"`
	Claps       uint      `gorm:"default:0" json:"claps"`
	Tags        []Tag     `gorm:"many2many:post_tags;" json:"tags"`
	Comment     uint      `gorm:"default:0" json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
