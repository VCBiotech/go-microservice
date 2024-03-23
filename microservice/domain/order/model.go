package order

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderID     uint64     `json:"order_id"     db:"order_id"`
	CustomerID  uuid.UUID  `json:"customer_id"  db:"customer_id"`
	LineItems   []LineItem `json:"line_items"   db:"line_items"`
	CreatedAt   *time.Time `json:"created_at"   db:"created_at"`
	ShippedAt   *time.Time `json:"shipped_at"   db:"shipped_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
}

type LineItem struct {
	ItemID   uuid.UUID `json:"item_id"  db:"item_id"`
	Quantity uint      `json:"quantity" db:"quantity"`
	Price    uint      `json:"price"    db:"price"`
}
