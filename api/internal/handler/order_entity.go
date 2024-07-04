package handler

import (
	"errors"
	"net/http"
)

type orderRequest struct {
	CustomerID string  `json:"customer_id"`
	Price      float64 `json:"price"`
}

func (o *orderRequest) Bind(_ *http.Request) error {
	if o.CustomerID == "" {
		return errors.New("customer_id is required")
	}
	return nil
}
