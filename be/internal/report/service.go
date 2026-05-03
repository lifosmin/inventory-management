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
	SKU         string  `json:"sku"`
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
		`SELECT l.id, l.lot_number, p.name, p.sku, l.quantity, l.unit_cost, l.total_cost
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
		if err := rows.Scan(&line.LotID, &line.LotNumber, &line.ProductName, &line.SKU,
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
