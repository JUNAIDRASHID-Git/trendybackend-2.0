package repository

import (
	"gorm.io/gorm"
	"trendybackend/internal/domain"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *domain.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) GetAll(search string, categoryID uint, subCategoryID uint) ([]domain.Product, error) {
	var products []domain.Product
	query := r.db.Model(&domain.Product{}).Preload("Tags").Preload("Variants.Images").Preload("Variants.Reviews")

	if search != "" {
		searchQuery := "%" + search + "%"
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(name_ar) LIKE LOWER(?)", 
			searchQuery, searchQuery)
	}

	if categoryID != 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	if subCategoryID != 0 {
		query = query.Where("sub_category_id = ?", subCategoryID)
	}

	err := query.Find(&products).Error
	return products, err
}

func (r *productRepository) GetByID(id uint) (*domain.Product, error) {
	var product domain.Product
	err := r.db.Preload("Tags").Preload("Variants.Images").Preload("Variants.Reviews").First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Update(product *domain.Product) error {
	// First save the product and its associations (like variants)
	if err := r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(product).Error; err != nil {
		return err
	}
	// Then replace associations for tags
	return r.db.Model(product).Association("Tags").Replace(product.Tags)
}

func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Product{}, id).Error
}

func (r *productRepository) updateParentStats(productID uint) {
	var product domain.Product
	if err := r.db.Preload("Variants.Images").First(&product, productID).Error; err == nil {
		r.db.Save(&product)
	}
}

func (r *productRepository) CreateVariant(variant *domain.ProductVariant) error {
	if err := r.db.Create(variant).Error; err != nil {
		return err
	}
	r.updateParentStats(variant.ProductID)
	return nil
}

func (r *productRepository) UpdateVariant(variant *domain.ProductVariant) error {
	if err := r.db.Save(variant).Error; err != nil {
		return err
	}
	// Replace images association if they are provided
	if len(variant.Images) > 0 {
		if err := r.db.Model(variant).Association("Images").Replace(variant.Images); err != nil {
			return err
		}
	}
	r.updateParentStats(variant.ProductID)
	return nil
}

func (r *productRepository) DeleteVariant(id uint) error {
	var variant domain.ProductVariant
	if err := r.db.First(&variant, id).Error; err != nil {
		return err
	}
	if err := r.db.Delete(&domain.ProductVariant{}, id).Error; err != nil {
		return err
	}
	r.updateParentStats(variant.ProductID)
	return nil
}

func (r *productRepository) GetVariantByID(id uint) (*domain.ProductVariant, error) {
	var variant domain.ProductVariant
	err := r.db.Preload("Images").Preload("Reviews").First(&variant, id).Error
	if err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *productRepository) AddVariantImage(img *domain.VariantImage) error {
	if err := r.db.Create(img).Error; err != nil {
		return err
	}
	// Find product ID from variant to update parent stats
	var variant domain.ProductVariant
	if err := r.db.First(&variant, img.VariantID).Error; err == nil {
		r.updateParentStats(variant.ProductID)
	}
	return nil
}

func (r *productRepository) DeleteVariantImage(id uint) error {
	var img domain.VariantImage
	if err := r.db.First(&img, id).Error; err != nil {
		return err
	}
	if err := r.db.Delete(&domain.VariantImage{}, id).Error; err != nil {
		return err
	}
	var variant domain.ProductVariant
	if err := r.db.First(&variant, img.VariantID).Error; err == nil {
		r.updateParentStats(variant.ProductID)
	}
	return nil
}

func (r *productRepository) AddVariantReview(rev *domain.VariantReview) error {
	if rev.CustomerName == "" && rev.CustomerID != 0 {
		var user domain.User
		if err := r.db.Select("first_name, last_name").First(&user, rev.CustomerID).Error; err == nil {
			rev.CustomerName = user.FirstName
			if user.LastName != "" {
				rev.CustomerName += " " + user.LastName
			}
		}
		if rev.CustomerName == "" {
			rev.CustomerName = "Customer"
		}
	}

	if err := r.db.Create(rev).Error; err != nil {
		return err
	}
	// Recalculate average rating and review count for the variant
	var stats struct {
		AvgRating float64
		Count     int64
	}
	r.db.Model(&domain.VariantReview{}).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(id) as count").
		Where("variant_id = ?", rev.VariantID).
		Scan(&stats)

	r.db.Model(&domain.ProductVariant{}).
		Where("id = ?", rev.VariantID).
		Updates(map[string]interface{}{
			"average_rating": stats.AvgRating,
			"review_count":   int(stats.Count),
		})

	return nil
}

func (r *productRepository) GetVariantReviews(variantID uint) ([]domain.VariantReview, error) {
	var reviews []domain.VariantReview
	err := r.db.Where("variant_id = ?", variantID).Order("created_at desc").Find(&reviews).Error
	return reviews, err
}
