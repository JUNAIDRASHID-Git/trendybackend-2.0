package usecase

import "trendybackend/internal/domain"

// ── Category Use Case ────────────────────────────────────────────────────────

type categoryUseCase struct {
	repo domain.CategoryRepository
}

func NewCategoryUseCase(repo domain.CategoryRepository) domain.CategoryUseCase {
	return &categoryUseCase{repo: repo}
}

func (uc *categoryUseCase) GetAllCategories() ([]domain.Category, error) {
	return uc.repo.GetAll()
}

func (uc *categoryUseCase) GetCategory(id uint) (*domain.Category, error) {
	return uc.repo.GetByID(id)
}

func (uc *categoryUseCase) CreateCategory(category *domain.Category) error {
	return uc.repo.Create(category)
}

func (uc *categoryUseCase) UpdateCategory(category *domain.Category) error {
	return uc.repo.Update(category)
}

func (uc *categoryUseCase) DeleteCategory(id uint) error {
	return uc.repo.Delete(id)
}

// ── SubCategory Use Case ─────────────────────────────────────────────────────

type subCategoryUseCase struct {
	repo domain.SubCategoryRepository
}

func NewSubCategoryUseCase(repo domain.SubCategoryRepository) domain.SubCategoryUseCase {
	return &subCategoryUseCase{repo: repo}
}

func (uc *subCategoryUseCase) GetAllSubCategories() ([]domain.SubCategory, error) {
	return uc.repo.GetAll()
}

func (uc *subCategoryUseCase) GetSubCategoriesByCategoryID(categoryID uint) ([]domain.SubCategory, error) {
	return uc.repo.GetByCategoryID(categoryID)
}

func (uc *subCategoryUseCase) GetSubCategory(id uint) (*domain.SubCategory, error) {
	return uc.repo.GetByID(id)
}

func (uc *subCategoryUseCase) CreateSubCategory(sub *domain.SubCategory) error {
	return uc.repo.Create(sub)
}

func (uc *subCategoryUseCase) UpdateSubCategory(sub *domain.SubCategory) error {
	return uc.repo.Update(sub)
}

func (uc *subCategoryUseCase) DeleteSubCategory(id uint) error {
	return uc.repo.Delete(id)
}
