package inventory

import (
	"time"
)

type MovementType string

const (
	MovementReceipt    MovementType = "receipt"
	MovementIssue      MovementType = "issue"
	MovementTransfer   MovementType = "transfer"
	MovementAdjustment MovementType = "adjustment"
	MovementReturn     MovementType = "return"
)

type Movement struct {
	ID           string       `json:"id"`
	LotID        string       `json:"lot_id"`
	MovementType MovementType `json:"movement_type"`
	Quantity     float64      `json:"quantity"`
	ReferenceDoc string       `json:"reference_doc,omitempty"`
	Notes        string       `json:"notes,omitempty"`
	PerformedBy  string       `json:"performed_by,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

type CreateMovementRequest struct {
	LotID        string       `json:"lot_id"`
	MovementType MovementType `json:"movement_type"`
	Quantity     float64      `json:"quantity"`
	ReferenceDoc string       `json:"reference_doc,omitempty"`
	Notes        string       `json:"notes,omitempty"`
}
