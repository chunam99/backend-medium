package models

import "time"

type Clap struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"not null"`
	PostID    uint `gorm:"not null"`
	Count     uint `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
