package repository

import (
	"trendybackend/internal/domain"

	"gorm.io/gorm"
)

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) domain.TagRepository {
	return &tagRepository{db}
}

func (r *tagRepository) Create(tag *domain.Tag) error {
	return r.db.Create(tag).Error
}

func (r *tagRepository) GetAll() ([]domain.Tag, error) {
	var tags []domain.Tag
	err := r.db.Find(&tags).Error
	return tags, err
}

func (r *tagRepository) GetByID(id uint) (*domain.Tag, error) {
	var tag domain.Tag
	err := r.db.First(&tag, id).Error
	return &tag, err
}

func (r *tagRepository) Update(tag *domain.Tag) error {
	return r.db.Save(tag).Error
}

func (r *tagRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Tag{}, id).Error
}
