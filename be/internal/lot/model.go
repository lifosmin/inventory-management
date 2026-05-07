package lot

import (
	"time"
)

type Status string

const (
	StatusAvailable Status = "available"
	StatusReserved  Status = "reserved"
	StatusDepleted  Status = "depleted"
	StatusExpired   Status = "expired"
)

type Lot struct {
	ID              string    `json:"id"`
	LotNumber       string    `json:"lot_number"`
	ProductID       string    `json:"product_id"`
	ProductName     string    `json:"product_name,omitempty"`
	WarehouseID     string    `json:"warehouse_id"`
	WarehouseName   string    `json:"warehouse_name,omitempty"`
	Quantity        float64   `json:"quantity"`
	InitialQuantity float64   `json:"initial_quantity"`
	UnitCost        float64   `json:"unit_cost"`
	TotalCost       float64   `json:"total_cost"`
	PaidAmount      float64   `json:"paid_amount"`
	PaymentStatus   string    `json:"payment_status"`
	ReceivedAt      time.Time `json:"received_at"`
	ExpiryDate      *string   `json:"expiry_date,omitempty"`
	Status          Status    `json:"status"`
	Supplier        string    `json:"supplier,omitempty"`
	ReferenceDoc    string    `json:"reference_doc,omitempty"`
	ShipmentStatus  string    `json:"shipment_status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateLotRequest struct {
	LotNumber    string  `json:"lot_number"`
	ProductID    string  `json:"product_id"`
	WarehouseID  string  `json:"warehouse_id"`
	Quantity     float64 `json:"quantity"`
	UnitCost     float64 `json:"unit_cost"`
	ExpiryDate   *string `json:"expiry_date,omitempty"`
	Supplier     string  `json:"supplier,omitempty"`
	ReferenceDoc string  `json:"reference_doc,omitempty"`
}

type UpdateLotRequest struct {
	Status      *Status  `json:"status,omitempty"`
	Quantity    *float64 `json:"quantity,omitempty"`
	ExpiryDate  *string  `json:"expiry_date,omitempty"`
	WarehouseID *string  `json:"warehouse_id,omitempty"`
}

type AddPaymentRequest struct {
	Amount float64 `json:"amount"`
}

type DeliverLotRequest struct {
	DeliveredDate     string  `json:"delivered_date"`
	AdditionalCost    float64 `json:"additional_cost"`
	ActualReceivedQty float64 `json:"actual_received_qty"`
}
