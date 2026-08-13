package usecase

import (
	"trendybackend/internal/domain"
)

type productUseCase struct {
	productRepo domain.ProductRepository
}

func NewProductUseCase(repo domain.ProductRepository) domain.ProductUseCase {
	return &productUseCase{
		productRepo: repo,
	}
}

func (u *productUseCase) GetAllProducts(search string, categoryID uint, subCategoryID uint) ([]domain.Product, error) {
	return u.productRepo.GetAll(search, categoryID, subCategoryID)
}

func (u *productUseCase) GetProduct(id uint) (*domain.Product, error) {
	return u.productRepo.GetByID(id)
}

func (u *productUseCase) CreateProduct(product *domain.Product) error {
	return u.productRepo.Create(product)
}

func (u *productUseCase) UpdateProduct(product *domain.Product) error {
	return u.productRepo.Update(product)
}

func (u *productUseCase) DeleteProduct(id uint) error {
	return u.productRepo.Delete(id)
}

func (u *productUseCase) CreateVariant(variant *domain.ProductVariant) error {
	return u.productRepo.CreateVariant(variant)
}

func (u *productUseCase) UpdateVariant(variant *domain.ProductVariant) error {
	return u.productRepo.UpdateVariant(variant)
}

func (u *productUseCase) DeleteVariant(id uint) error {
	return u.productRepo.DeleteVariant(id)
}

func (u *productUseCase) GetVariantByID(id uint) (*domain.ProductVariant, error) {
	return u.productRepo.GetVariantByID(id)
}

func (u *productUseCase) AddVariantImage(img *domain.VariantImage) error {
	return u.productRepo.AddVariantImage(img)
}

func (u *productUseCase) DeleteVariantImage(id uint) error {
	return u.productRepo.DeleteVariantImage(id)
}

func (u *productUseCase) AddVariantReview(rev *domain.VariantReview) error {
	return u.productRepo.AddVariantReview(rev)
}

func (u *productUseCase) GetVariantReviews(variantID uint) ([]domain.VariantReview, error) {
	return u.productRepo.GetVariantReviews(variantID)
}
