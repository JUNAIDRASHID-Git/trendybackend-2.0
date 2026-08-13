package domain

import (
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null;unique" json:"name"`
	NameAr    string         `json:"name_ar"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type TagRepository interface {
	Create(tag *Tag) error
	GetAll() ([]Tag, error)
	GetByID(id uint) (*Tag, error)
	Update(tag *Tag) error
	Delete(id uint) error
}

type TagUseCase interface {
	GetAllTags() ([]Tag, error)
	GetTag(id uint) (*Tag, error)
	CreateTag(tag *Tag) error
	UpdateTag(tag *Tag) error
	DeleteTag(id uint) error
}
