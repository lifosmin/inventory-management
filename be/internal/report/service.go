package report

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type ValuationLine struct {
	LotID       string  `json:"lot_id"`
	LotNumber   string  `json:"lot_number"`
	ProductName string  `json:"product_name"`
	Quantity    float64 `json:"quantity"`
	UnitCost    float64 `json:"unit_cost"`
	TotalCost   float64 `json:"total_cost"`
}

type ValuationReport struct {
	Lines      []ValuationLine `json:"lines"`
	TotalValue float64         `json:"total_value"`
}

func (s *Service) InventoryValuation(ctx context.Context) (*ValuationReport, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT l.id, l.lot_number, p.name, l.quantity, l.unit_cost, l.total_cost
		 FROM lots l JOIN products p ON l.product_id = p.id
		 WHERE l.status = 'available' AND l.quantity > 0
		 ORDER BY p.name, l.received_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying valuation: %w", err)
	}
	defer rows.Close()

	report := &ValuationReport{}
	for rows.Next() {
		var line ValuationLine
		if err := rows.Scan(&line.LotID, &line.LotNumber, &line.ProductName,
			&line.Quantity, &line.UnitCost, &line.TotalCost); err != nil {
			return nil, fmt.Errorf("scanning valuation line: %w", err)
		}
		report.TotalValue += line.TotalCost
		report.Lines = append(report.Lines, line)
	}
	return report, nil
}

type AgingLine struct {
	LotID       string    `json:"lot_id"`
	LotNumber   string    `json:"lot_number"`
	ProductName string    `json:"product_name"`
	Quantity    float64   `json:"quantity"`
	ReceivedAt  time.Time `json:"received_at"`
	DaysOld     int       `json:"days_old"`
}

type AgingReport struct {
	Lines []AgingLine `json:"lines"`
}

func (s *Service) LotAging(ctx context.Context) (*AgingReport, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT l.id, l.lot_number, p.name, l.quantity, l.received_at,
		        EXTRACT(DAY FROM now() - l.received_at)::int AS days_old
		 FROM lots l JOIN products p ON l.product_id = p.id
		 WHERE l.status = 'available' AND l.quantity > 0
		 ORDER BY l.received_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying aging: %w", err)
	}
	defer rows.Close()

	report := &AgingReport{}
	for rows.Next() {
		var line AgingLine
		if err := rows.Scan(&line.LotID, &line.LotNumber, &line.ProductName,
			&line.Quantity, &line.ReceivedAt, &line.DaysOld); err != nil {
			return nil, fmt.Errorf("scanning aging line: %w", err)
		}
		report.Lines = append(report.Lines, line)
	}
	return report, nil
}

type StockLine struct {
	ProductName      string  `json:"product_name"`
	WarehouseName    string  `json:"warehouse_name"`
	TotalQty         float64 `json:"total_qty"`
	TotalInitialQty  float64 `json:"total_initial_qty"`
	TotalValue       float64 `json:"total_value"`
}

type RestockLine struct {
	LotNumber       string    `json:"lot_number"`
	ProductName     string    `json:"product_name"`
	WarehouseName   string    `json:"warehouse_name"`
	Quantity        float64   `json:"quantity"`
	InitialQuantity float64   `json:"initial_quantity"`
	UnitCost        float64   `json:"unit_cost"`
	TotalCost       float64   `json:"total_cost"`
	Supplier        string    `json:"supplier"`
	CreatedAt       time.Time `json:"created_at"`
}

type SaleLine struct {
	BuyerName     string    `json:"buyer_name"`
	ProductName   string    `json:"product_name"`
	WarehouseName string    `json:"warehouse_name"`
	Qty           float64   `json:"qty"`
	SellPrice     float64   `json:"sell_price"`
	Total         float64   `json:"total"`
	CreatedAt     time.Time `json:"created_at"`
}

type ExportReport struct {
	From          string        `json:"from"`
	To            string        `json:"to"`
	NetWorth      float64       `json:"net_worth"`
	TotalRevenue  float64       `json:"total_revenue"`
	TotalCOGS     float64       `json:"total_cogs"`
	ProfitLoss    float64       `json:"profit_loss"`
	Stocks        []StockLine   `json:"stocks"`
	Restocks      []RestockLine `json:"restocks"`
	Sales         []SaleLine    `json:"sales"`
}

func (s *Service) Export(ctx context.Context, from, to time.Time) (*ExportReport, error) {
	report := &ExportReport{
		From: from.Format("2006-01-02"),
		To:   to.Format("2006-01-02"),
	}

	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(quantity * unit_cost), 0)
		 FROM lots WHERE status = 'available' AND quantity > 0`,
	).Scan(&report.NetWorth)
	if err != nil {
		return nil, fmt.Errorf("querying net worth: %w", err)
	}

	err = s.pool.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(sa.qty * s.sell_price), 0),
			COALESCE(SUM(sa.qty * sa.unit_cost), 0)
		 FROM sale_allocations sa
		 JOIN sales s ON sa.sale_id = s.id
		 WHERE s.created_at >= $1 AND s.created_at < $2`,
		from, to.Add(24*time.Hour),
	).Scan(&report.TotalRevenue, &report.TotalCOGS)
	if err != nil {
		return nil, fmt.Errorf("querying revenue: %w", err)
	}
	report.ProfitLoss = report.TotalRevenue - report.TotalCOGS

	stockRows, err := s.pool.Query(ctx,
		`SELECT p.name, w.name,
		        COALESCE(SUM(l.quantity), 0),
		        COALESCE(SUM(l.initial_quantity), 0),
		        COALESCE(SUM(l.quantity * l.unit_cost), 0)
		 FROM lots l
		 JOIN products p ON l.product_id = p.id
		 JOIN warehouses w ON l.warehouse_id = w.id
		 WHERE l.status = 'available' AND l.quantity > 0
		 GROUP BY p.name, w.name
		 ORDER BY p.name, w.name`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying stocks: %w", err)
	}
	defer stockRows.Close()
	for stockRows.Next() {
		var sl StockLine
		if err := stockRows.Scan(&sl.ProductName, &sl.WarehouseName, &sl.TotalQty, &sl.TotalInitialQty, &sl.TotalValue); err != nil {
			return nil, fmt.Errorf("scanning stock: %w", err)
		}
		report.Stocks = append(report.Stocks, sl)
	}

	restockRows, err := s.pool.Query(ctx,
		`SELECT l.lot_number, p.name, w.name, l.quantity, l.initial_quantity, l.unit_cost, l.total_cost, COALESCE(l.supplier, ''), l.created_at
		 FROM lots l
		 JOIN products p ON l.product_id = p.id
		 JOIN warehouses w ON l.warehouse_id = w.id
		 WHERE l.created_at >= $1 AND l.created_at < $2
		 ORDER BY l.created_at DESC`,
		from, to.Add(24*time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("querying restocks: %w", err)
	}
	defer restockRows.Close()
	for restockRows.Next() {
		var rl RestockLine
		if err := restockRows.Scan(&rl.LotNumber, &rl.ProductName, &rl.WarehouseName,
			&rl.Quantity, &rl.InitialQuantity, &rl.UnitCost, &rl.TotalCost, &rl.Supplier, &rl.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning restock: %w", err)
		}
		report.Restocks = append(report.Restocks, rl)
	}

	saleRows, err := s.pool.Query(ctx,
		`SELECT s.buyer_name, p.name, w.name, s.qty, s.sell_price, s.qty * s.sell_price, s.created_at
		 FROM sales s
		 JOIN products p ON s.product_id = p.id
		 JOIN warehouses w ON s.warehouse_id = w.id
		 WHERE s.created_at >= $1 AND s.created_at < $2
		 ORDER BY s.created_at DESC`,
		from, to.Add(24*time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("querying sales: %w", err)
	}
	defer saleRows.Close()
	for saleRows.Next() {
		var sl SaleLine
		if err := saleRows.Scan(&sl.BuyerName, &sl.ProductName, &sl.WarehouseName,
			&sl.Qty, &sl.SellPrice, &sl.Total, &sl.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning sale: %w", err)
		}
		report.Sales = append(report.Sales, sl)
	}

	return report, nil
}
