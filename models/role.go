package models

import "time"

type Role struct {
	ID        uint   `gorm:"primaryKey"`
	RoleName  string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
