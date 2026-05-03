package product

import (
	"time"
)

type Product struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateProductRequest struct {
	Name     string `json:"name"`
	Category string `json:"category,omitempty"`
}

type UpdateProductRequest struct {
	Name     *string `json:"name,omitempty"`
	Category *string `json:"category,omitempty"`
}
