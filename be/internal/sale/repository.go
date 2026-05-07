package sale

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateTx(ctx context.Context, tx pgx.Tx, req CreateSaleRequest) (*Sale, error) {
	var s Sale
	err := tx.QueryRow(ctx,
		`INSERT INTO sales (product_id, warehouse_id, buyer_name, qty, sell_price)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, product_id, warehouse_id, buyer_name, qty, sell_price, paid_amount, payment_status, shipment_status, created_at`,
		req.ProductID, req.WarehouseID, req.BuyerName, req.Qty, req.SellPrice,
	).Scan(&s.ID, &s.ProductID, &s.WarehouseID, &s.BuyerName, &s.Qty, &s.SellPrice,
		&s.PaidAmount, &s.PaymentStatus, &s.ShipmentStatus, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating sale: %w", err)
	}
	return &s, nil
}

func (r *Repository) CreateAllocationTx(ctx context.Context, tx pgx.Tx, saleID, lotID string, qty, unitCost float64) (*Allocation, error) {
	var a Allocation
	err := tx.QueryRow(ctx,
		`INSERT INTO sale_allocations (sale_id, lot_id, qty, unit_cost)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, sale_id, lot_id, qty, unit_cost, created_at`,
		saleID, lotID, qty, unitCost,
	).Scan(&a.ID, &a.SaleID, &a.LotID, &a.Qty, &a.UnitCost, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating allocation: %w", err)
	}
	return &a, nil
}

func (r *Repository) List(ctx context.Context, params ListParams) (*ListResult, error) {
	allowedSort := map[string]string{
		"buyer_name":      "s.buyer_name",
		"product_name":    "p.name",
		"warehouse_name":  "w.name",
		"qty":             "s.qty",
		"sell_price":      "s.sell_price",
		"total":           "s.qty * s.sell_price",
		"paid_amount":     "s.paid_amount",
		"payment_status":  "s.payment_status",
		"shipment_status": "s.shipment_status",
		"created_at":      "s.created_at",
	}
	sortCol, ok := allowedSort[params.SortBy]
	if !ok {
		sortCol = "s.created_at"
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
		where = append(where, fmt.Sprintf("(s.buyer_name ILIKE %s OR p.name ILIKE %s)", p, p))
	}
	if params.ProductID != "" {
		where = append(where, fmt.Sprintf("s.product_id = %s", nextArg(params.ProductID)))
	}
	if params.WarehouseID != "" {
		where = append(where, fmt.Sprintf("s.warehouse_id = %s", nextArg(params.WarehouseID)))
	}
	if params.ShipmentStatus != "" {
		where = append(where, fmt.Sprintf("s.shipment_status = %s", nextArg(params.ShipmentStatus)))
	}
	if params.PaymentStatus != "" {
		where = append(where, fmt.Sprintf("s.payment_status = %s", nextArg(params.PaymentStatus)))
	}
	if params.DateFrom != "" {
		where = append(where, fmt.Sprintf("s.created_at >= %s::date", nextArg(params.DateFrom)))
	}
	if params.DateTo != "" {
		where = append(where, fmt.Sprintf("s.created_at < (%s::date + interval '1 day')", nextArg(params.DateTo)))
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	countArgs := append([]any{}, args...)
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*)
		 FROM sales s
		 JOIN products p ON s.product_id = p.id
		 JOIN warehouses w ON s.warehouse_id = w.id
		 `+whereSQL, countArgs...,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("counting sales: %w", err)
	}

	limitArg := nextArg(params.Limit)
	offsetArg := nextArg(params.Offset)

	rows, err := r.pool.Query(ctx,
		`SELECT s.id, s.product_id, p.name, s.warehouse_id, w.name,
		        s.buyer_name, s.qty, s.sell_price, s.paid_amount, s.payment_status, s.shipment_status, s.created_at
		 FROM sales s
		 JOIN products p ON s.product_id = p.id
		 JOIN warehouses w ON s.warehouse_id = w.id
		 `+whereSQL+`
		 ORDER BY `+sortCol+` `+sortDir+`
		 LIMIT `+limitArg+` OFFSET `+offsetArg, args...,
	)
	if err != nil {
		return nil, fmt.Errorf("listing sales: %w", err)
	}
	defer rows.Close()

	var sales []Sale
	for rows.Next() {
		var s Sale
		if err := rows.Scan(&s.ID, &s.ProductID, &s.ProductName, &s.WarehouseID, &s.WarehouseName,
			&s.BuyerName, &s.Qty, &s.SellPrice, &s.PaidAmount, &s.PaymentStatus, &s.ShipmentStatus, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning sale: %w", err)
		}
		sales = append(sales, s)
	}
	if sales == nil {
		sales = []Sale{}
	}
	return &ListResult{Total: total, Data: sales}, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sales SET
			shipment_status = COALESCE($2, shipment_status)
		 WHERE id = $1`,
		id, req.ShipmentStatus,
	)
	if err != nil {
		return fmt.Errorf("updating sale status: %w", err)
	}
	return nil
}

func (r *Repository) AddPayment(ctx context.Context, id string, amount float64) (*Sale, error) {
	var currentPaid, totalOwed float64
	err := r.pool.QueryRow(ctx,
		`SELECT paid_amount, qty * sell_price FROM sales WHERE id = $1`, id,
	).Scan(&currentPaid, &totalOwed)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fetching sale for payment: %w", err)
	}
	if currentPaid+amount > totalOwed {
		return nil, ErrPaymentExceedsTotal
	}

	var s Sale
	err = r.pool.QueryRow(ctx,
		`UPDATE sales SET
			paid_amount = paid_amount + $2,
			payment_status = CASE
				WHEN paid_amount + $2 <= 0 THEN 'unpaid'
				WHEN paid_amount + $2 >= qty * sell_price THEN 'fully_paid'
				ELSE 'dp'
			END
		 WHERE id = $1
		 RETURNING id, product_id, warehouse_id, buyer_name, qty, sell_price, paid_amount, payment_status, shipment_status, created_at`,
		id, amount,
	).Scan(&s.ID, &s.ProductID, &s.WarehouseID, &s.BuyerName, &s.Qty, &s.SellPrice,
		&s.PaidAmount, &s.PaymentStatus, &s.ShipmentStatus, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("adding sale payment: %w", err)
	}
	return &s, nil
}
func (r *Repository) Cancel(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var shipmentStatus string
	err = tx.QueryRow(ctx,
		`SELECT shipment_status FROM sales WHERE id = $1 FOR UPDATE`, id,
	).Scan(&shipmentStatus)
	if err == pgx.ErrNoRows {
		return ErrSaleNotFound
	}
	if err != nil {
		return fmt.Errorf("fetching sale for cancel: %w", err)
	}
	if shipmentStatus == "delivered" {
		return ErrSaleAlreadyDelivered
	}
	if shipmentStatus == "canceled" {
		return ErrSaleAlreadyCanceled
	}

	rows, err := tx.Query(ctx,
		`SELECT lot_id, qty FROM sale_allocations WHERE sale_id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("fetching allocations: %w", err)
	}
	type alloc struct {
		lotID string
		qty   float64
	}
	var allocs []alloc
	for rows.Next() {
		var a alloc
		if err := rows.Scan(&a.lotID, &a.qty); err != nil {
			rows.Close()
			return fmt.Errorf("scanning allocation: %w", err)
		}
		allocs = append(allocs, a)
	}
	rows.Close()

	for _, a := range allocs {
		_, err = tx.Exec(ctx,
			`UPDATE lots SET
				quantity   = quantity + $2,
				status     = CASE WHEN status != 'canceled' THEN 'available' ELSE status END,
				updated_at = now()
			 WHERE id = $1`, a.lotID, a.qty,
		)
		if err != nil {
			return fmt.Errorf("restoring lot quantity: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `DELETE FROM sale_allocations WHERE sale_id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting allocations: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE sales SET shipment_status = 'canceled' WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("canceling sale: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing sale cancel: %w", err)
	}
	return nil
}
