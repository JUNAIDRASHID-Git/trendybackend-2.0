package domain

import "time"

type ProductActivity struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID string    `gorm:"type:varchar(255);index;column:product_id" json:"product_id"` // zoho_item_id
	Type      string    `gorm:"type:varchar(50);index;column:type" json:"type"`             // "view", "wishlist", "purchase"
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
