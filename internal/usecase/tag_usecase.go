package usecase

import (
	"trendybackend/internal/domain"
)

type tagUseCase struct {
	tagRepo domain.TagRepository
}

func NewTagUseCase(tagRepo domain.TagRepository) domain.TagUseCase {
	return &tagUseCase{
		tagRepo: tagRepo,
	}
}

func (u *tagUseCase) CreateTag(tag *domain.Tag) error {
	return u.tagRepo.Create(tag)
}

func (u *tagUseCase) GetAllTags() ([]domain.Tag, error) {
	return u.tagRepo.GetAll()
}

func (u *tagUseCase) GetTag(id uint) (*domain.Tag, error) {
	return u.tagRepo.GetByID(id)
}

func (u *tagUseCase) UpdateTag(tag *domain.Tag) error {
	return u.tagRepo.Update(tag)
}

func (u *tagUseCase) DeleteTag(id uint) error {
	return u.tagRepo.Delete(id)
}
