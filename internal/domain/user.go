package domain

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Email           string         `gorm:"uniqueIndex;not null" json:"email"`
	Password        string         `gorm:"not null" json:"-"`
	FirstName       string         `json:"first_name"`
	LastName        string         `json:"last_name"`
	Phone           string         `json:"phone"`
	ProfileImageUrl string         `json:"profile_image_url"`
	Role            string         `gorm:"default:'admin'" json:"role"` // 'super_admin', 'admin', 'customer'
	AddressCity     string         `json:"address_city"`
	AddressDistrict string         `json:"address_district"`
	AddressLandmark string         `json:"address_landmark"`
	AddressLat      float64        `json:"address_lat"`
	AddressLon      float64        `json:"address_lon"`
	CartItems       string         `json:"cart_items"` // JSON serialized cart items
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type CustomerWithStats struct {
	User
	TotalOrders int     `json:"total_orders"`
	TotalSpent  float64 `json:"total_spent"`
}

type UserRepository interface {
	Create(user *User) error
	Update(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id uint) (*User, error)
	FindAll() ([]User, error)
	Delete(id uint) error
	GetCustomersWithStats() ([]CustomerWithStats, error)
}
