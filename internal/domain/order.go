package domain

import (
	"time"
)

type Order struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserID          uint      `json:"userId" gorm:"column:user_id;index"`
	CustomerName    string    `json:"customerName" binding:"required"`
	CustomerEmail   string    `json:"customerEmail"`
	CustomerPhone   string    `json:"customerPhone"`
	CustomerAddress string    `json:"customerAddress"`
	PaymentMethod   string    `json:"paymentMethod"`
	OrderType       string    `json:"orderType"`
	ItemsJson       string    `json:"itemsJson" gorm:"column:items_json"` // JSON: [{name, qty, price, imageUrl}]
	TotalAmount     float64   `json:"totalAmount" binding:"required"`
	Status          string    `json:"status" binding:"required"` // Pending, Preparing, OutForDelivery, Delivered, Cancelled
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
