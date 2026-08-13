package repository

import (
	"gorm.io/gorm"
	"trendybackend/internal/domain"
)

// ── Category Repository ───────────────────────────────────────────────────────

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *domain.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) GetAll() ([]domain.Category, error) {
	var categories []domain.Category
	err := r.db.Preload("SubCategories").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetByID(id uint) (*domain.Category, error) {
	var category domain.Category
	err := r.db.Preload("SubCategories").First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Update(category *domain.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Category{}, id).Error
}

// ── SubCategory Repository ───────────────────────────────────────────────────

type subCategoryRepository struct {
	db *gorm.DB
}

func NewSubCategoryRepository(db *gorm.DB) domain.SubCategoryRepository {
	return &subCategoryRepository{db: db}
}

func (r *subCategoryRepository) Create(sub *domain.SubCategory) error {
	return r.db.Create(sub).Error
}

func (r *subCategoryRepository) GetAll() ([]domain.SubCategory, error) {
	var subs []domain.SubCategory
	err := r.db.Preload("Category").Find(&subs).Error
	return subs, err
}

func (r *subCategoryRepository) GetByCategoryID(categoryID uint) ([]domain.SubCategory, error) {
	var subs []domain.SubCategory
	err := r.db.Where("category_id = ?", categoryID).Find(&subs).Error
	return subs, err
}

func (r *subCategoryRepository) GetByID(id uint) (*domain.SubCategory, error) {
	var sub domain.SubCategory
	err := r.db.Preload("Category").First(&sub, id).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subCategoryRepository) Update(sub *domain.SubCategory) error {
	return r.db.Save(sub).Error
}

func (r *subCategoryRepository) Delete(id uint) error {
	return r.db.Delete(&domain.SubCategory{}, id).Error
}
