package lot

import (
	"context"
	"fmt"
	"strings"
	"time"

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

func (r *Repository) List(ctx context.Context, params ListParams) (*ListResult, error) {
	allowedSort := map[string]string{
		"lot_number":      "l.lot_number",
		"product_name":    "p.name",
		"warehouse_name":  "w.name",
		"quantity":        "l.quantity",
		"unit_cost":       "l.unit_cost",
		"total_cost":      "l.total_cost",
		"paid_amount":     "l.paid_amount",
		"payment_status":  "l.payment_status",
		"shipment_status": "l.shipment_status",
		"received_at":     "l.received_at",
		"created_at":      "l.created_at",
	}
	sortCol, ok := allowedSort[params.SortBy]
	if !ok {
		sortCol = "l.received_at"
	}
	sortDir := "DESC"
	if params.SortDir == "asc" {
		sortDir = "ASC"
	}
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	args := []any{}
	where := []string{}
	nextArg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if params.Search != "" {
		p := nextArg("%" + params.Search + "%")
		where = append(where, fmt.Sprintf("(l.lot_number ILIKE %s OR p.name ILIKE %s OR l.supplier ILIKE %s)", p, p, p))
	}
	if params.ProductID != "" {
		where = append(where, fmt.Sprintf("l.product_id = %s", nextArg(params.ProductID)))
	}
	if params.WarehouseID != "" {
		where = append(where, fmt.Sprintf("l.warehouse_id = %s", nextArg(params.WarehouseID)))
	}
	if params.ShipmentStatus != "" {
		where = append(where, fmt.Sprintf("l.shipment_status = %s", nextArg(params.ShipmentStatus)))
	}
	if params.PaymentStatus != "" {
		where = append(where, fmt.Sprintf("l.payment_status = %s", nextArg(params.PaymentStatus)))
	}
	if params.DateFrom != "" {
		where = append(where, fmt.Sprintf("l.created_at >= %s::date", nextArg(params.DateFrom)))
	}
	if params.DateTo != "" {
		where = append(where, fmt.Sprintf("l.created_at < (%s::date + interval '1 day')", nextArg(params.DateTo)))
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	countArgs := append([]any{}, args...)
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*)
		 FROM lots l
		 JOIN products p ON l.product_id = p.id
		 JOIN warehouses w ON l.warehouse_id = w.id
		 `+whereSQL, countArgs...,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("counting lots: %w", err)
	}

	limitArg := nextArg(params.Limit)
	offsetArg := nextArg(params.Offset)

	rows, err := r.pool.Query(ctx,
		`SELECT l.id, l.lot_number, l.product_id, p.name, l.warehouse_id, w.name,
		        l.quantity, l.initial_quantity, l.unit_cost, l.total_cost,
		        l.paid_amount, l.payment_status, l.received_at, l.expiry_date, l.status, l.supplier, l.reference_doc,
		        l.shipment_status, l.created_at, l.updated_at
		 FROM lots l
		 JOIN products p ON l.product_id = p.id
		 JOIN warehouses w ON l.warehouse_id = w.id
		 `+whereSQL+`
		 ORDER BY `+sortCol+` `+sortDir+`
		 LIMIT `+limitArg+` OFFSET `+offsetArg, args...,
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
	if lots == nil {
		lots = []Lot{}
	}
	return &ListResult{Total: total, Data: lots}, nil
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
	var strategy string
	err := r.pool.QueryRow(ctx,
		`SELECT fifo_strategy FROM warehouses WHERE id = $1`, warehouseID,
	).Scan(&strategy)
	if err != nil {
		strategy = "created_at"
	}

	var orderBy string
	var deliveredOnly bool
	switch strategy {
	case "received_at":
		orderBy = "received_at ASC"
		deliveredOnly = true
	default:
		orderBy = "created_at ASC"
		deliveredOnly = false
	}

	var rows pgx.Rows
	if deliveredOnly {
		rows, err = r.pool.Query(ctx,
			`SELECT id, lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost, total_cost,
			        received_at, expiry_date, status, supplier, reference_doc, shipment_status, created_at, updated_at
			 FROM lots
			 WHERE product_id = $1 AND warehouse_id = $2 AND status = 'available' AND quantity > 0
			   AND shipment_status = 'delivered'
			 ORDER BY `+orderBy, productID, warehouseID,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost, total_cost,
			        received_at, expiry_date, status, supplier, reference_doc, shipment_status, created_at, updated_at
			 FROM lots
			 WHERE product_id = $1 AND warehouse_id = $2 AND status = 'available' AND quantity > 0
			 ORDER BY `+orderBy, productID, warehouseID,
		)
	}
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

func (r *Repository) MarkDelivered(ctx context.Context, id string, deliveredDate string, additionalCost float64, actualReceivedQty float64) (*Lot, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var origInitialQty, origQty, origUnitCost, origPaidAmount float64
	var origLotNumber, origProductID, origWarehouseID, origSupplier, origReferenceDoc, origShipmentStatus string
	var origExpiryDate *string
	err = tx.QueryRow(ctx,
		`SELECT lot_number, product_id, warehouse_id, initial_quantity, quantity, unit_cost,
		        paid_amount, supplier, reference_doc, expiry_date, shipment_status
		 FROM lots WHERE id = $1 FOR UPDATE`, id,
	).Scan(&origLotNumber, &origProductID, &origWarehouseID, &origInitialQty, &origQty,
		&origUnitCost, &origPaidAmount, &origSupplier, &origReferenceDoc, &origExpiryDate, &origShipmentStatus)
	if err == pgx.ErrNoRows {
		return nil, ErrLotNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("fetching lot for delivery: %w", err)
	}

	if origShipmentStatus == "delivered" {
		return nil, ErrAlreadyDelivered
	}

	if actualReceivedQty > origInitialQty {
		return nil, ErrQtyExceedsOrdered
	}

	alreadySold := origInitialQty - origQty
	if actualReceivedQty < alreadySold {
		return nil, ErrActualQtyBelowSold
	}

	newUnitCost := (actualReceivedQty*origUnitCost + additionalCost) / actualReceivedQty
	newQty := actualReceivedQty - alreadySold

	var l Lot
	err = tx.QueryRow(ctx,
		`UPDATE lots SET
			shipment_status  = 'delivered',
			initial_quantity = $2,
			quantity         = $3,
			unit_cost        = $4,
			received_at      = $5::date,
			updated_at       = now()
		 WHERE id = $1
		 RETURNING id, lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost, total_cost,
		           paid_amount, payment_status, received_at, expiry_date, status, supplier, reference_doc, shipment_status, created_at, updated_at`,
		id, actualReceivedQty, newQty, newUnitCost, deliveredDate,
	).Scan(&l.ID, &l.LotNumber, &l.ProductID, &l.WarehouseID, &l.Quantity, &l.InitialQuantity, &l.UnitCost, &l.TotalCost,
		&l.PaidAmount, &l.PaymentStatus, &l.ReceivedAt, &l.ExpiryDate, &l.Status, &l.Supplier, &l.ReferenceDoc, &l.ShipmentStatus, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("updating lot for delivery: %w", err)
	}

	remainingQty := origInitialQty - actualReceivedQty
	if remainingQty > 0 {
		// Split paid_amount proportionally between delivered and back-order portions
		boPaidAmount := origPaidAmount * (remainingQty / origInitialQty)
		origNewPaid := origPaidAmount - boPaidAmount

		// Determine payment_status for the back-order lot
		boCost := remainingQty * origUnitCost
		var boPaymentStatus string
		if boPaidAmount <= 0 {
			boPaymentStatus = "unpaid"
		} else if boPaidAmount >= boCost {
			boPaymentStatus = "fully_paid"
		} else {
			boPaymentStatus = "dp"
		}

		// Determine updated payment_status for the original lot
		origCost := actualReceivedQty * newUnitCost
		var origPaymentStatus string
		if origNewPaid <= 0 {
			origPaymentStatus = "unpaid"
		} else if origNewPaid >= origCost {
			origPaymentStatus = "fully_paid"
		} else {
			origPaymentStatus = "dp"
		}

		// Update the original lot's paid_amount and payment_status to reflect its smaller share
		_, err = tx.Exec(ctx,
			`UPDATE lots SET paid_amount = $2, payment_status = $3, updated_at = now() WHERE id = $1`,
			id, origNewPaid, origPaymentStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("updating original lot payment after split: %w", err)
		}

		var boCount int
		_ = tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM lots WHERE lot_number LIKE $1`,
			origLotNumber+"-BO%",
		).Scan(&boCount)
		var boLotNumber string
		if boCount == 0 {
			boLotNumber = origLotNumber + "-BO"
		} else {
			boLotNumber = fmt.Sprintf("%s-BO%d", origLotNumber, boCount+1)
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO lots (lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost,
			                   paid_amount, payment_status, supplier, reference_doc, expiry_date,
			                   shipment_status, status, received_at)
			 VALUES ($1, $2, $3, $4, $4, $5, $6, $7, $8, $9, $10, 'in_progress', 'available', $11)`,
			boLotNumber, origProductID, origWarehouseID, remainingQty, origUnitCost,
			boPaidAmount, boPaymentStatus, origSupplier, origReferenceDoc, origExpiryDate, time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("creating back-order lot: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing delivery transaction: %w", err)
	}
	return &l, nil
}

func (r *Repository) Cancel(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var shipmentStatus string
	err = tx.QueryRow(ctx,
		`SELECT shipment_status FROM lots WHERE id = $1 FOR UPDATE`, id,
	).Scan(&shipmentStatus)
	if err == pgx.ErrNoRows {
		return ErrLotNotFound
	}
	if err != nil {
		return fmt.Errorf("fetching lot for cancel: %w", err)
	}
	if shipmentStatus == "delivered" {
		return ErrAlreadyDelivered
	}
	if shipmentStatus == "canceled" {
		return ErrLotAlreadyCanceled
	}

	var allocCount int
	err = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM sale_allocations sa
		 JOIN sales s ON sa.sale_id = s.id
		 WHERE sa.lot_id = $1 AND s.shipment_status != 'canceled'`, id,
	).Scan(&allocCount)
	if err != nil {
		return fmt.Errorf("checking dependent sales: %w", err)
	}
	if allocCount > 0 {
		return ErrLotHasDependentSales
	}

	_, err = tx.Exec(ctx,
		`UPDATE lots SET shipment_status = 'canceled', status = 'depleted', updated_at = now() WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("canceling lot: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing lot cancel: %w", err)
	}
	return nil
}

type AvailableQtyItem struct {
	WarehouseID   string  `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	Qty           float64 `json:"qty"`
}

func (r *Repository) GetAvailableQtyByProduct(ctx context.Context, productID string) ([]AvailableQtyItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT w.id, w.name, COALESCE(SUM(l.quantity), 0)
		 FROM warehouses w
		 LEFT JOIN lots l ON l.warehouse_id = w.id
		   AND l.product_id = $1
		   AND l.status = 'available'
		   AND l.quantity > 0
		 GROUP BY w.id, w.name
		 ORDER BY w.name`, productID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying available qty: %w", err)
	}
	defer rows.Close()
	var items []AvailableQtyItem
	for rows.Next() {
		var item AvailableQtyItem
		if err := rows.Scan(&item.WarehouseID, &item.WarehouseName, &item.Qty); err != nil {
			return nil, fmt.Errorf("scanning available qty: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}
