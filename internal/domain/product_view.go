package domain

import "time"

type ProductView struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID string    `gorm:"type:varchar(255);index;column:product_id" json:"product_id"`
	UserID    *uint     `gorm:"index;column:user_id" json:"user_id,omitempty"`
	GuestID   string    `gorm:"type:varchar(255);index;column:guest_id" json:"guest_id"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
