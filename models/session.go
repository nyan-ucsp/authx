package models

import "time"

type Session struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserId       uint      `gorm:"foreignKey:UserId" json:"userId"`
	User         *User     `gorm:"foreignKey:UserId"`
	RefreshToken string    `json:"refreshToken"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	ExpiredAt    time.Time `json:"expiredAt"`
}
