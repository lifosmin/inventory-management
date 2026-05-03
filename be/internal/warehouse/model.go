package warehouse

import (
	"time"
)

type Warehouse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateWarehouseRequest struct {
	Name    string `json:"name"`
	Address string `json:"address,omitempty"`
}

type UpdateWarehouseRequest struct {
	Name    *string `json:"name,omitempty"`
	Address *string `json:"address,omitempty"`
}
