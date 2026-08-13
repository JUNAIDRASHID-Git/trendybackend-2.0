package domain

import "time"

// ZohoProduct represents a product synchronized from Zoho Books
type ZohoProduct struct {
	ZohoItemID          string              `gorm:"primaryKey;uniqueIndex;column:zoho_item_id;type:varchar(255)" json:"zoho_item_id"`
	Name                string              `gorm:"type:varchar(255);not null" json:"name"`
	NameAr              string              `gorm:"type:varchar(255)" json:"name_ar"`
	Description         string              `gorm:"type:text" json:"description"`
	DescriptionAr       string              `gorm:"type:text" json:"description_ar"`
	Rate                float64             `gorm:"type:decimal(10,2);not null" json:"rate"`
	SKU                 string              `gorm:"type:varchar(255)" json:"sku"`
	Stock               int                 `gorm:"type:integer;default:0" json:"stock"`
	Weight              float64             `gorm:"type:decimal(10,2);default:0" json:"weight"`
	CategoryID          uint                `json:"category_id"`
	SubCategoryID       uint                `json:"sub_category_id"`
	CustomImage         string              `gorm:"type:varchar(255)" json:"custom_image"`
	IsVisibleToCustomer bool                `gorm:"column:is_visible_to_customer;default:true" json:"is_visible_to_customer"`
	IsTrending          bool                `gorm:"column:is_trending;default:false" json:"is_trending"`
	IsRecommended       bool                `gorm:"column:is_recommended;default:false" json:"is_recommended"`
	SalesVolume         int                 `gorm:"column:sales_volume;default:0" json:"sales_volume"`
	Reviews             []ZohoProductReview `gorm:"foreignKey:ZohoProductID;constraint:OnDelete:CASCADE" json:"reviews,omitempty"`
	AverageRating       float64             `gorm:"-" json:"average_rating"`
	ReviewCount         int                 `gorm:"-" json:"review_count"`
	MetaTitle           string              `gorm:"type:varchar(255)" json:"meta_title"`
	MetaDescription     string              `gorm:"type:text" json:"meta_description"`
	Keywords            string              `gorm:"type:text" json:"keywords"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
}
