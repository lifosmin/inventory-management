package sale

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lifosmin/admin-backend/internal/lot"
)

var (
	ErrInsufficientStock    = errors.New("insufficient stock for this sale")
	ErrPaymentExceedsTotal  = errors.New("payment exceeds total owed")
	ErrSaleNotFound         = errors.New("sale not found")
	ErrSaleAlreadyDelivered = errors.New("sale is already delivered")
	ErrSaleAlreadyCanceled  = errors.New("sale is already canceled")
)

type Service struct {
	repo    *Repository
	lotRepo *lot.Repository
	pool    *pgxpool.Pool
}

func NewService(repo *Repository, lotRepo *lot.Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, lotRepo: lotRepo, pool: pool}
}

func (s *Service) Create(ctx context.Context, req CreateSaleRequest) (*Sale, error) {
	lots, err := s.lotRepo.ListAvailableByProductWarehouse(ctx, req.ProductID, req.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("fetching available lots: %w", err)
	}

	var totalAvailable float64
	for _, l := range lots {
		totalAvailable += l.Quantity
	}
	if totalAvailable < req.Qty {
		return nil, ErrInsufficientStock
	}

	var sale *Sale
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	sale, err = s.repo.CreateTx(ctx, tx, req)
	if err != nil {
		return nil, err
	}

	remaining := req.Qty
	var allocations []Allocation
	for _, l := range lots {
		if remaining <= 0 {
			break
		}
		allocQty := l.Quantity
		if allocQty > remaining {
			allocQty = remaining
		}

		alloc, err := s.repo.CreateAllocationTx(ctx, tx, sale.ID, l.ID, allocQty, l.UnitCost)
		if err != nil {
			return nil, err
		}
		allocations = append(allocations, *alloc)

		newQty := l.Quantity - allocQty
		status := lot.StatusAvailable
		if newQty <= 0 {
			status = lot.StatusDepleted
			newQty = 0
		}
		_, err = tx.Exec(ctx,
			`UPDATE lots SET quantity = $2, status = $3, updated_at = now() WHERE id = $1`,
			l.ID, newQty, status,
		)
		if err != nil {
			return nil, fmt.Errorf("updating lot quantity: %w", err)
		}

		remaining -= allocQty
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing sale: %w", err)
	}

	sale.Allocations = allocations
	return sale, nil
}

func (s *Service) List(ctx context.Context) ([]Sale, error) {
	return s.repo.List(ctx)
}

func (s *Service) UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) error {
	return s.repo.UpdateStatus(ctx, id, req)
}

func (s *Service) AddPayment(ctx context.Context, id string, amount float64) (*Sale, error) {
	return s.repo.AddPayment(ctx, id, amount)
}

func (s *Service) Cancel(ctx context.Context, id string) error {
	return s.repo.Cancel(ctx, id)
}

func (s *Service) GetDashboard(ctx context.Context) (*Dashboard, error) {
	var d Dashboard

	// Realized P&L — only non-canceled sales
	err := s.pool.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(sa.qty * s.sell_price), 0),
			COALESCE(SUM(sa.qty * sa.unit_cost), 0)
		 FROM sale_allocations sa
		 JOIN sales s ON sa.sale_id = s.id
		 WHERE s.shipment_status != 'canceled'`,
	).Scan(&d.TotalRevenue, &d.TotalCOGS)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("querying revenue/cogs: %w", err)
	}
	d.ProfitLoss = d.TotalRevenue - d.TotalCOGS

	// Pending revenue — outstanding balance on active unpaid/dp sales
	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(qty * sell_price - paid_amount), 0)
		 FROM sales
		 WHERE payment_status != 'fully_paid' AND shipment_status != 'canceled'`,
	).Scan(&d.PendingRevenue)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("querying pending revenue: %w", err)
	}

	// Stock on hand — delivered lots still available
	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(quantity * unit_cost), 0)
		 FROM lots
		 WHERE status = 'available' AND shipment_status = 'delivered' AND quantity > 0`,
	).Scan(&d.StockValueOnHand)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("querying stock on hand: %w", err)
	}

	// Stock in transit — restocks not yet received
	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(initial_quantity * unit_cost), 0)
		 FROM lots WHERE shipment_status = 'in_progress'`,
	).Scan(&d.StockValueInTransit)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("querying stock in transit: %w", err)
	}

	// Lot payment status counts (restocks)
	err = s.pool.QueryRow(ctx,
		`SELECT
			COUNT(*) FILTER (WHERE payment_status = 'unpaid'),
			COUNT(*) FILTER (WHERE payment_status = 'dp')
		 FROM lots WHERE shipment_status != 'canceled'`,
	).Scan(&d.LotsUnpaidCount, &d.LotsDPCount)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("querying lot payment counts: %w", err)
	}

	// Stock summary per product (only products with active stock)
	stockRows, err := s.pool.Query(ctx,
		`SELECT
			p.name,
			COALESCE(SUM(CASE WHEN l.shipment_status = 'delivered' AND l.status = 'available' THEN l.quantity ELSE 0 END), 0) AS qty_on_hand,
			COALESCE(SUM(CASE WHEN l.shipment_status = 'in_progress' THEN l.initial_quantity ELSE 0 END), 0) AS qty_in_transit
		 FROM products p
		 LEFT JOIN lots l ON l.product_id = p.id AND l.shipment_status != 'canceled'
		 GROUP BY p.id, p.name
		 HAVING
			SUM(CASE WHEN l.shipment_status = 'delivered' AND l.status = 'available' THEN l.quantity ELSE 0 END) > 0
			OR SUM(CASE WHEN l.shipment_status = 'in_progress' THEN l.initial_quantity ELSE 0 END) > 0
		 ORDER BY p.name`,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("querying stock summary: %w", err)
	}
	d.StockSummary = []StockSummaryItem{}
	if stockRows != nil {
		defer stockRows.Close()
		for stockRows.Next() {
			var item StockSummaryItem
			if err := stockRows.Scan(&item.ProductName, &item.QtyOnHand, &item.QtyInTransit); err != nil {
				return nil, fmt.Errorf("scanning stock summary: %w", err)
			}
			d.StockSummary = append(d.StockSummary, item)
		}
	}

	// Active sales — not delivered and not canceled
	saleRows, err := s.pool.Query(ctx,
		`SELECT s.id, s.buyer_name, p.name, s.qty, s.sell_price, s.paid_amount, s.payment_status, s.shipment_status, s.created_at
		 FROM sales s
		 JOIN products p ON s.product_id = p.id
		 WHERE s.shipment_status NOT IN ('delivered', 'canceled')
		 ORDER BY s.created_at DESC`,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("querying active sales: %w", err)
	}
	d.ActiveSales = []ActiveSaleItem{}
	if saleRows != nil {
		defer saleRows.Close()
		for saleRows.Next() {
			var item ActiveSaleItem
			if err := saleRows.Scan(&item.ID, &item.BuyerName, &item.ProductName, &item.Qty, &item.SellPrice,
				&item.PaidAmount, &item.PaymentStatus, &item.ShipmentStatus, &item.CreatedAt); err != nil {
				return nil, fmt.Errorf("scanning active sale: %w", err)
			}
			d.ActiveSales = append(d.ActiveSales, item)
		}
	}

	// Buyers with outstanding balance
	debtRows, err := s.pool.Query(ctx,
		`SELECT buyer_name, SUM(qty * sell_price - paid_amount) AS amount_owed
		 FROM sales
		 WHERE payment_status != 'fully_paid' AND shipment_status != 'canceled'
		 GROUP BY buyer_name
		 ORDER BY amount_owed DESC`,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("querying buyers owing: %w", err)
	}
	d.BuyersOwing = []TopCustomer{}
	if debtRows != nil {
		defer debtRows.Close()
		for debtRows.Next() {
			var c TopCustomer
			if err := debtRows.Scan(&c.BuyerName, &c.AmountOwed); err != nil {
				return nil, fmt.Errorf("scanning buyer debt: %w", err)
			}
			d.BuyersOwing = append(d.BuyersOwing, c)
		}
	}

	return &d, nil
}

type TopCustomer struct {
	BuyerName  string  `json:"buyer_name"`
	AmountOwed float64 `json:"amount_owed"`
}

type StockSummaryItem struct {
	ProductName  string  `json:"product_name"`
	QtyOnHand    float64 `json:"qty_on_hand"`
	QtyInTransit float64 `json:"qty_in_transit"`
}

type ActiveSaleItem struct {
	ID             string    `json:"id"`
	BuyerName      string    `json:"buyer_name"`
	ProductName    string    `json:"product_name"`
	Qty            float64   `json:"qty"`
	SellPrice      float64   `json:"sell_price"`
	PaidAmount     float64   `json:"paid_amount"`
	PaymentStatus  string    `json:"payment_status"`
	ShipmentStatus string    `json:"shipment_status"`
	CreatedAt      time.Time `json:"created_at"`
}

type Dashboard struct {
	TotalRevenue        float64            `json:"total_revenue"`
	TotalCOGS           float64            `json:"total_cogs"`
	ProfitLoss          float64            `json:"profit_loss"`
	PendingRevenue      float64            `json:"pending_revenue"`
	StockValueOnHand    float64            `json:"stock_value_on_hand"`
	StockValueInTransit float64            `json:"stock_value_in_transit"`
	LotsUnpaidCount     int                `json:"lots_unpaid_count"`
	LotsDPCount         int                `json:"lots_dp_count"`
	StockSummary        []StockSummaryItem `json:"stock_summary"`
	ActiveSales         []ActiveSaleItem   `json:"active_sales"`
	BuyersOwing         []TopCustomer      `json:"buyers_owing"`
}
