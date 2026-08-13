package repository

import (
	"strings"
	"trendybackend/internal/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *domain.User) error {
	// db.Save() performs a full row UPDATE including zero-value fields (e.g. lat/lon = 0.0).
	// This is correct since we always fetch the full user struct before calling Update.
	return r.db.Save(user).Error
}


func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll() ([]domain.User, error) {
	var users []domain.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}

func (r *userRepository) GetCustomersWithStats() ([]domain.CustomerWithStats, error) {
	var customers []domain.CustomerWithStats

	var users []domain.User
	if err := r.db.Where("role = ?", "customer").Find(&users).Error; err != nil {
		return nil, err
	}

	for _, u := range users {
		var totalOrders int64
		var totalSpent float64

		email := strings.ToLower(strings.TrimSpace(u.Email))

		// Strict user account matching by User ID or unique Customer Email (no name collisions)
		r.db.Model(&domain.Order{}).Where(
			"user_id = ? OR (LOWER(customer_email) = ? AND customer_email != '')",
			u.ID, email,
		).Count(&totalOrders)

		// Sum total_amount for non-cancelled orders of this specific user account
		r.db.Model(&domain.Order{}).Where(
			"(user_id = ? OR (LOWER(customer_email) = ? AND customer_email != '')) AND LOWER(status) != ?",
			u.ID, email, "cancelled",
		).Select("COALESCE(SUM(total_amount), 0)").Row().Scan(&totalSpent)

		customers = append(customers, domain.CustomerWithStats{
			User:        u,
			TotalOrders: int(totalOrders),
			TotalSpent:  totalSpent,
		})
	}

	return customers, nil
}
