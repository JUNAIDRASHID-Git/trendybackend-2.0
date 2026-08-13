package usecase

import (
	"trendybackend/internal/domain"
	"trendybackend/internal/repository"
)

type OrderUseCase interface {
	GetAll() ([]domain.Order, error)
	GetByID(id uint) (domain.Order, error)
	GetByEmail(email string) ([]domain.Order, error)
	Create(order *domain.Order) error
	Update(order *domain.Order) error
	Delete(id uint) error
}

type orderUseCase struct {
	repo repository.OrderRepository
}

func NewOrderUseCase(repo repository.OrderRepository) OrderUseCase {
	return &orderUseCase{repo}
}

func (u *orderUseCase) GetAll() ([]domain.Order, error) {
	return u.repo.GetAll()
}

func (u *orderUseCase) GetByID(id uint) (domain.Order, error) {
	return u.repo.GetByID(id)
}

func (u *orderUseCase) GetByEmail(email string) ([]domain.Order, error) {
	return u.repo.GetByEmail(email)
}

func (u *orderUseCase) Create(order *domain.Order) error {
	return u.repo.Create(order)
}

func (u *orderUseCase) Update(order *domain.Order) error {
	return u.repo.Update(order)
}

func (u *orderUseCase) Delete(id uint) error {
	return u.repo.Delete(id)
}
