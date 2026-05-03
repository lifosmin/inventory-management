package product

import (
	"time"
)

type Product struct {
	ID        string    `json:"id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Unit      string    `json:"unit"`
	Category  string    `json:"category,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateProductRequest struct {
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Unit     string `json:"unit"`
	Category string `json:"category,omitempty"`
}

type UpdateProductRequest struct {
	Name     *string `json:"name,omitempty"`
	Unit     *string `json:"unit,omitempty"`
	Category *string `json:"category,omitempty"`
}
