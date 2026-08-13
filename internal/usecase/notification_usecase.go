package usecase

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"gorm.io/gorm"
	"trendybackend/internal/domain"
)

type NotificationUseCase interface {
	GetNotifications() ([]domain.Notification, error)
}

type notificationUseCase struct {
	db *gorm.DB
}

func NewNotificationUseCase(db *gorm.DB) NotificationUseCase {
	return &notificationUseCase{db: db}
}

func (u *notificationUseCase) GetNotifications() ([]domain.Notification, error) {
	var notifications []domain.Notification

	// 1. Low Stock Alerts (Products with stock < 5)
	var lowStockProducts []domain.Product
	if err := u.db.Where("stock < ?", 5).Find(&lowStockProducts).Error; err == nil {
		for _, p := range lowStockProducts {
			notifications = append(notifications, domain.Notification{
				ID:        "low_stock_" + strconv.FormatUint(uint64(p.ID), 10),
				Type:      "low_stock",
				Title:     "Low Stock Alert",
				Message:   fmt.Sprintf("Product '%s' is running low on stock (%d left)", p.Name, p.Stock),
				CreatedAt: p.UpdatedAt,
			})
		}
	}

	// 2. New Orders (Orders created within the last 24 hours)
	var newOrders []domain.Order
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	if err := u.db.Where("created_at >= ?", oneDayAgo).Find(&newOrders).Error; err == nil {
		for _, o := range newOrders {
			notifications = append(notifications, domain.Notification{
				ID:        "new_order_" + strconv.FormatUint(uint64(o.ID), 10),
				Type:      "new_order",
				Title:     "New Order Placed",
				Message:   fmt.Sprintf("Order #%d for SAR %.2f was placed by %s", o.ID, o.TotalAmount, o.CustomerName),
				CreatedAt: o.CreatedAt,
			})
		}
	}

	// 3. Delayed Pending Orders (Orders in Pending status created > 24 hours ago)
	var delayedOrders []domain.Order
	if err := u.db.Where("status = ? AND created_at < ?", "Pending", oneDayAgo).Find(&delayedOrders).Error; err == nil {
		for _, o := range delayedOrders {
			notifications = append(notifications, domain.Notification{
				ID:        "pending_delay_" + strconv.FormatUint(uint64(o.ID), 10),
				Type:      "pending_delay",
				Title:     "Delayed Pending Order",
				Message:   fmt.Sprintf("Order #%d has been pending for over 24 hours.", o.ID),
				CreatedAt: o.CreatedAt,
			})
		}
	}

	// Sort aggregated notifications by CreatedAt descending (newest first)
	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].CreatedAt.After(notifications[j].CreatedAt)
	})

	return notifications, nil
}
