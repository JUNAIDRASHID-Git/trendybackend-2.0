package repository

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"trendybackend/internal/domain"
)

type OrderRepository interface {
	GetAll() ([]domain.Order, error)
	GetByID(id uint) (domain.Order, error)
	GetByEmail(email string) ([]domain.Order, error)
	Create(order *domain.Order) error
	Update(order *domain.Order) error
	Delete(id uint) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) GetAll() ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetByID(id uint) (domain.Order, error) {
	var order domain.Order
	err := r.db.First(&order, id).Error
	return order, err
}

func (r *orderRepository) GetByEmail(email string) ([]domain.Order, error) {
	var orders []domain.Order
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	if cleanEmail == "" {
		return orders, nil
	}

	// Find corresponding user profile if exists
	var user domain.User
	if err := r.db.Where("LOWER(email) = ?", cleanEmail).First(&user).Error; err == nil {
		err := r.db.Where(
			"user_id = ? OR (LOWER(customer_email) = ? AND customer_email != '')",
			user.ID, cleanEmail,
		).Order("created_at DESC").Find(&orders).Error
		return orders, err
	}

	// Fallback for guest checkout using exact email only
	err := r.db.Where("LOWER(customer_email) = ?", cleanEmail).Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) Create(order *domain.Order) error {
	// Auto-bind user_id if customer email matches an existing user account
	if order.UserID == 0 && strings.TrimSpace(order.CustomerEmail) != "" {
		var user domain.User
		if err := r.db.Where("LOWER(email) = ?", strings.ToLower(strings.TrimSpace(order.CustomerEmail))).First(&user).Error; err == nil {
			order.UserID = user.ID
		}
	}

	err := r.db.Create(order).Error
	if err == nil {
		var items []struct {
			ID         int    `json:"id"`
			ZohoItemID string `json:"zoho_item_id"`
			ProductID  int    `json:"productId"`
		}
		if errJson := json.Unmarshal([]byte(order.ItemsJson), &items); errJson == nil {
			for _, item := range items {
				var prodID string
				if item.ZohoItemID != "" {
					prodID = item.ZohoItemID
				} else if item.ID > 0 {
					prodID = strconv.Itoa(item.ID)
				} else if item.ProductID > 0 {
					prodID = strconv.Itoa(item.ProductID)
				}
				if prodID != "" {
					activity := domain.ProductActivity{
						ProductID: prodID,
						Type:      "purchase",
						CreatedAt: time.Now(),
					}
					r.db.Create(&activity)
				}
			}
		}
	}
	return err
}

func (r *orderRepository) Update(order *domain.Order) error {
	return r.db.Save(order).Error
}

func (r *orderRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Order{}, id).Error
}
