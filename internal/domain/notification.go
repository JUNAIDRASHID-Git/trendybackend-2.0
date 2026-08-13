package domain

import "time"

type Notification struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "low_stock", "new_order", "pending_delay"
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}
