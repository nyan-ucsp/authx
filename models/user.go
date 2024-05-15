package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`             // Standard field for the primary key
	UserName  string    `gorm:"type:varchar(50)" json:"username"` // A regular string field
	Email     *string   `gorm:"type:varchar(255)" json:"email"`   // A pointer to a string, allowing for null values
	Phone     *string   `gorm:"type:varchar(50)" json:"phone"`    // A pointer to a string, allowing for null values
	Password  string    `json:"-"`                                // A regular string field
	CreatedAt time.Time `json:"createdAt"`                        // Automatically managed by GORM for creation time
	UpdatedAt time.Time `json:"updatedAt"`                        // Automatically managed by GORM for update time
}
