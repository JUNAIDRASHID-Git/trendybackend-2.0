package domain

import "time"

// ZohoProductReview represents customer reviews left on synced Zoho products
type ZohoProductReview struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ZohoProductID string    `gorm:"index;type:varchar(255);not null" json:"zoho_product_id"`
	CustomerID    uint      `gorm:"index" json:"customer_id"`
	CustomerName  string    `gorm:"type:varchar(255)" json:"customer_name"`
	Rating        int       `json:"rating"`
	Review        string    `json:"review"`
	CreatedAt     time.Time `json:"created_at"`
}
