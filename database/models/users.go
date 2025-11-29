package models

import "time"

type User struct {
	ID         string `gorm:"primaryKey"`
	Name       string
	Email      string  `gorm:"unique"`
	CustomerID *string `gorm:"size:255"`
	Plan       string  `gorm:"size:50;default:'free';not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
