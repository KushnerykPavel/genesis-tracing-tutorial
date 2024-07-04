package entity

import "time"

type Order struct {
	CustomerID string
	OrderID    int
	Price      float64
	CreatedAt  time.Time
}
