package model

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderID     uint64     `json:"order_id"`
	CustomerID  uuid.UUID  `json:"customer_id"`
	LineItems   []LineItem `json:"line_items"`
	OrderStatus string     `json:"order_status"`
	CreatedAt   *time.Time `json:"created_at`
	UpdatedAt   *time.Time `json:"updated_at`
	CompletedAt *time.Time `json:"completed_at`
	ShippedAt   *time.Time `json:"shipped_at`
}

type LineItem struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity uint      `json:"quantity"`
	Price    uint      `json:"price"`
}
