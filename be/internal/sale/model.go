package sale

import (
	"time"
)

type Sale struct {
	ID             string       `json:"id"`
	ProductID      string       `json:"product_id"`
	ProductName    string       `json:"product_name,omitempty"`
	WarehouseID    string       `json:"warehouse_id"`
	WarehouseName  string       `json:"warehouse_name,omitempty"`
	BuyerName      string       `json:"buyer_name"`
	Qty            float64      `json:"qty"`
	SellPrice      float64      `json:"sell_price"`
	PaidAmount     float64      `json:"paid_amount"`
	PaymentStatus  string       `json:"payment_status"`
	ShipmentStatus string       `json:"shipment_status"`
	CreatedAt      time.Time    `json:"created_at"`
	Allocations    []Allocation `json:"allocations,omitempty"`
}

type Allocation struct {
	ID        string    `json:"id"`
	SaleID    string    `json:"sale_id"`
	LotID     string    `json:"lot_id"`
	LotNumber string    `json:"lot_number,omitempty"`
	Qty       float64   `json:"qty"`
	UnitCost  float64   `json:"unit_cost"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateSaleRequest struct {
	ProductID   string  `json:"product_id"`
	WarehouseID string  `json:"warehouse_id"`
	BuyerName   string  `json:"buyer_name"`
	Qty         float64 `json:"qty"`
	SellPrice   float64 `json:"sell_price"`
}

type UpdateStatusRequest struct {
	ShipmentStatus *string `json:"shipment_status,omitempty"`
}

type AddPaymentRequest struct {
	Amount float64 `json:"amount"`
}
