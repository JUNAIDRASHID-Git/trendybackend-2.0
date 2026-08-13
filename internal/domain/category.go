package domain

import (
	"gorm.io/gorm"
	"time"
)

// Category is the top-level product grouping (e.g. "Desserts", "Beverages")
type Category struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	Name          string          `gorm:"not null;uniqueIndex" json:"name"`
	Description   string          `json:"description"`
	ImageURL      string          `json:"image_url"`
	SubCategories []SubCategory   `gorm:"foreignKey:CategoryID" json:"sub_categories,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`
}

// SubCategory belongs to a Category (e.g. "Chocolate Cakes" under "Desserts")
type SubCategory struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"not null" json:"name"`
	Description string        `json:"description"`
	CategoryID uint           `gorm:"not null;index" json:"category_id"`
	Category   *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// Repository interfaces
type CategoryRepository interface {
	Create(category *Category) error
	GetAll() ([]Category, error)
	GetByID(id uint) (*Category, error)
	Update(category *Category) error
	Delete(id uint) error
}

type SubCategoryRepository interface {
	Create(sub *SubCategory) error
	GetAll() ([]SubCategory, error)
	GetByCategoryID(categoryID uint) ([]SubCategory, error)
	GetByID(id uint) (*SubCategory, error)
	Update(sub *SubCategory) error
	Delete(id uint) error
}

// Use case interfaces
type CategoryUseCase interface {
	GetAllCategories() ([]Category, error)
	GetCategory(id uint) (*Category, error)
	CreateCategory(category *Category) error
	UpdateCategory(category *Category) error
	DeleteCategory(id uint) error
}

type SubCategoryUseCase interface {
	GetAllSubCategories() ([]SubCategory, error)
	GetSubCategoriesByCategoryID(categoryID uint) ([]SubCategory, error)
	GetSubCategory(id uint) (*SubCategory, error)
	CreateSubCategory(sub *SubCategory) error
	UpdateSubCategory(sub *SubCategory) error
	DeleteSubCategory(id uint) error
}
