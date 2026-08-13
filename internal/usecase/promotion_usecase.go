package usecase

import "trendybackend/internal/domain"

type promotionUseCase struct {
	repo domain.PromotionRepository
}

func NewPromotionUseCase(repo domain.PromotionRepository) domain.PromotionUseCase {
	return &promotionUseCase{repo: repo}
}

func (u *promotionUseCase) CreatePromotion(promotion *domain.Promotion) error {
	return u.repo.Create(promotion)
}

func (u *promotionUseCase) GetAllPromotions() ([]domain.Promotion, error) {
	return u.repo.GetAll()
}

func (u *promotionUseCase) GetPromotion(id uint) (*domain.Promotion, error) {
	return u.repo.GetByID(id)
}

func (u *promotionUseCase) UpdatePromotion(promotion *domain.Promotion) error {
	return u.repo.Update(promotion)
}

func (u *promotionUseCase) DeletePromotion(id uint) error {
	return u.repo.Delete(id)
}
