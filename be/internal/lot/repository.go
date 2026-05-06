package lot

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, req CreateLotRequest) (*Lot, error) {
	var l Lot
	err := r.pool.QueryRow(ctx,
		`INSERT INTO lots (lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost, expiry_date, supplier, reference_doc)
		 VALUES ($1, $2, $3, $4, $4, $5, $6, $7, $8)
		 RETURNING id, lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost, total_cost,
		           paid_amount, payment_status, received_at, expiry_date, status, supplier, reference_doc, shipment_status, created_at, updated_at`,
		req.LotNumber, req.ProductID, req.WarehouseID, req.Quantity, req.UnitCost,
		req.ExpiryDate, req.Supplier, req.ReferenceDoc,
	).Scan(&l.ID, &l.LotNumber, &l.ProductID, &l.WarehouseID, &l.Quantity, &l.InitialQuantity, &l.UnitCost, &l.TotalCost,
		&l.PaidAmount, &l.PaymentStatus, &l.ReceivedAt, &l.ExpiryDate, &l.Status, &l.Supplier, &l.ReferenceDoc, &l.ShipmentStatus, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating lot: %w", err)
	}
	return &l, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Lot, error) {
	var l Lot
	err := r.pool.QueryRow(ctx,
		`SELECT id, lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost, total_cost,
		        paid_amount, payment_status, received_at, expiry_date, status, supplier, reference_doc, shipment_status, created_at, updated_at
		 FROM lots WHERE id = $1`, id,
	).Scan(&l.ID, &l.LotNumber, &l.ProductID, &l.WarehouseID, &l.Quantity, &l.InitialQuantity, &l.UnitCost, &l.TotalCost,
		&l.PaidAmount, &l.PaymentStatus, &l.ReceivedAt, &l.ExpiryDate, &l.Status, &l.Supplier, &l.ReferenceDoc, &l.ShipmentStatus, &l.CreatedAt, &l.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting lot: %w", err)
	}
	return &l, nil
}

func (r *Repository) List(ctx context.Context) ([]Lot, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT l.id, l.lot_number, l.product_id, p.name, l.warehouse_id, w.name,
		        l.quantity, l.initial_quantity, l.unit_cost, l.total_cost,
		        l.paid_amount, l.payment_status, l.received_at, l.expiry_date, l.status, l.supplier, l.reference_doc,
		        l.shipment_status, l.created_at, l.updated_at
		 FROM lots l
		 JOIN products p ON l.product_id = p.id
		 JOIN warehouses w ON l.warehouse_id = w.id
		 ORDER BY l.received_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing lots: %w", err)
	}
	defer rows.Close()

	var lots []Lot
	for rows.Next() {
		var l Lot
		if err := rows.Scan(&l.ID, &l.LotNumber, &l.ProductID, &l.ProductName, &l.WarehouseID, &l.WarehouseName,
			&l.Quantity, &l.InitialQuantity, &l.UnitCost, &l.TotalCost,
			&l.PaidAmount, &l.PaymentStatus, &l.ReceivedAt, &l.ExpiryDate, &l.Status, &l.Supplier, &l.ReferenceDoc,
			&l.ShipmentStatus, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning lot: %w", err)
		}
		lots = append(lots, l)
	}
	return lots, nil
}

func (r *Repository) ListByProductFIFO(ctx context.Context, productID string) ([]Lot, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost, total_cost,
		        received_at, expiry_date, status, supplier, reference_doc, shipment_status, created_at, updated_at
		 FROM lots
		 WHERE product_id = $1 AND status = 'available' AND quantity > 0
		 ORDER BY received_at ASC`, productID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing lots by product FIFO: %w", err)
	}
	defer rows.Close()

	var lots []Lot
	for rows.Next() {
		var l Lot
		if err := rows.Scan(&l.ID, &l.LotNumber, &l.ProductID, &l.WarehouseID, &l.Quantity, &l.InitialQuantity, &l.UnitCost, &l.TotalCost,
			&l.ReceivedAt, &l.ExpiryDate, &l.Status, &l.Supplier, &l.ReferenceDoc, &l.ShipmentStatus, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning lot: %w", err)
		}
		lots = append(lots, l)
	}
	return lots, nil
}

func (r *Repository) ListAvailableByProductWarehouse(ctx context.Context, productID, warehouseID string) ([]Lot, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost, total_cost,
		        received_at, expiry_date, status, supplier, reference_doc, shipment_status, created_at, updated_at
		 FROM lots
		 WHERE product_id = $1 AND warehouse_id = $2 AND status = 'available' AND quantity > 0
		 ORDER BY received_at ASC`, productID, warehouseID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing available lots: %w", err)
	}
	defer rows.Close()

	var lots []Lot
	for rows.Next() {
		var l Lot
		if err := rows.Scan(&l.ID, &l.LotNumber, &l.ProductID, &l.WarehouseID, &l.Quantity, &l.InitialQuantity, &l.UnitCost, &l.TotalCost,
			&l.ReceivedAt, &l.ExpiryDate, &l.Status, &l.Supplier, &l.ReferenceDoc, &l.ShipmentStatus, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning lot: %w", err)
		}
		lots = append(lots, l)
	}
	return lots, nil
}

func (r *Repository) Update(ctx context.Context, id string, req UpdateLotRequest) (*Lot, error) {
	var l Lot
	err := r.pool.QueryRow(ctx,
		`UPDATE lots SET
			status = COALESCE($2, status),
			quantity = COALESCE($3, quantity),
			expiry_date = COALESCE($4, expiry_date),
			warehouse_id = COALESCE($5, warehouse_id),
			updated_at = now()
		 WHERE id = $1
		 RETURNING id, lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost, total_cost,
		           paid_amount, payment_status, received_at, expiry_date, status, supplier, reference_doc, shipment_status, created_at, updated_at`,
		id, req.Status, req.Quantity, req.ExpiryDate, req.WarehouseID,
	).Scan(&l.ID, &l.LotNumber, &l.ProductID, &l.WarehouseID, &l.Quantity, &l.InitialQuantity, &l.UnitCost, &l.TotalCost,
		&l.PaidAmount, &l.PaymentStatus, &l.ReceivedAt, &l.ExpiryDate, &l.Status, &l.Supplier, &l.ReferenceDoc, &l.ShipmentStatus, &l.CreatedAt, &l.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("updating lot: %w", err)
	}
	return &l, nil
}

func (r *Repository) UpdateShipmentStatus(ctx context.Context, id string, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE lots SET shipment_status = $2, updated_at = now() WHERE id = $1`,
		id, status,
	)
	if err != nil {
		return fmt.Errorf("updating shipment status: %w", err)
	}
	return nil
}

func (r *Repository) UpdateQuantity(ctx context.Context, id string, newQty float64, status Status) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE lots SET quantity = $2, status = $3, updated_at = now() WHERE id = $1`,
		id, newQty, status,
	)
	if err != nil {
		return fmt.Errorf("updating lot quantity: %w", err)
	}
	return nil
}

func (r *Repository) AddPayment(ctx context.Context, id string, amount float64) (*Lot, error) {
	var currentPaid, totalOwed float64
	err := r.pool.QueryRow(ctx,
		`SELECT paid_amount, initial_quantity * unit_cost FROM lots WHERE id = $1`, id,
	).Scan(&currentPaid, &totalOwed)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fetching lot for payment: %w", err)
	}
	if currentPaid+amount > totalOwed {
		return nil, ErrPaymentExceedsTotal
	}

	var l Lot
	err = r.pool.QueryRow(ctx,
		`UPDATE lots SET
			paid_amount = paid_amount + $2,
			payment_status = CASE
				WHEN paid_amount + $2 <= 0 THEN 'unpaid'
				WHEN paid_amount + $2 >= initial_quantity * unit_cost THEN 'fully_paid'
				ELSE 'dp'
			END,
			updated_at = now()
		 WHERE id = $1
		 RETURNING id, lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost, total_cost,
		           paid_amount, payment_status, received_at, expiry_date, status, supplier, reference_doc, shipment_status, created_at, updated_at`,
		id, amount,
	).Scan(&l.ID, &l.LotNumber, &l.ProductID, &l.WarehouseID, &l.Quantity, &l.InitialQuantity, &l.UnitCost, &l.TotalCost,
		&l.PaidAmount, &l.PaymentStatus, &l.ReceivedAt, &l.ExpiryDate, &l.Status, &l.Supplier, &l.ReferenceDoc, &l.ShipmentStatus, &l.CreatedAt, &l.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("adding lot payment: %w", err)
	}
	return &l, nil
}
