package repository

import (
	"gorm.io/gorm"
	"trendybackend/internal/domain"
)

type promotionRepository struct {
	db *gorm.DB
}

func NewPromotionRepository(db *gorm.DB) domain.PromotionRepository {
	return &promotionRepository{db: db}
}

func (r *promotionRepository) Create(promotion *domain.Promotion) error {
	if len(promotion.ProductIDs) > 0 {
		var products []domain.ZohoProduct
		r.db.Where("zoho_item_id IN ?", promotion.ProductIDs).Find(&products)
		promotion.Products = products
	}
	return r.db.Create(promotion).Error
}

func (r *promotionRepository) GetAll() ([]domain.Promotion, error) {
	var promotions []domain.Promotion
	err := r.db.Preload("Products").Find(&promotions).Error
	return promotions, err
}

func (r *promotionRepository) GetByID(id uint) (*domain.Promotion, error) {
	var promotion domain.Promotion
	err := r.db.Preload("Products").First(&promotion, id).Error
	return &promotion, err
}

func (r *promotionRepository) Update(promotion *domain.Promotion) error {
	if len(promotion.ProductIDs) > 0 {
		var products []domain.ZohoProduct
		r.db.Where("zoho_item_id IN ?", promotion.ProductIDs).Find(&products)
		r.db.Model(promotion).Association("Products").Replace(products)
	} else {
		r.db.Model(promotion).Association("Products").Clear()
	}
	return r.db.Save(promotion).Error
}

func (r *promotionRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Promotion{}, id).Error
}
