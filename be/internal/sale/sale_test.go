package sale_test

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lifosmin/admin-backend/internal/lot"
	"github.com/lifosmin/admin-backend/internal/sale"
	"github.com/lifosmin/admin-backend/internal/testutil"
)

type fixtures struct {
	productID   string
	warehouseID string
}

func seedFixtures(t *testing.T, pool *pgxpool.Pool) fixtures {
	t.Helper()
	ctx := context.Background()

	var f fixtures
	err := pool.QueryRow(ctx,
		`INSERT INTO products (name, category) VALUES ('Test Product', 'test') RETURNING id`,
	).Scan(&f.productID)
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}
	err = pool.QueryRow(ctx,
		`INSERT INTO warehouses (name, address) VALUES ('Test WH', 'addr') RETURNING id`,
	).Scan(&f.warehouseID)
	if err != nil {
		t.Fatalf("seed warehouse: %v", err)
	}
	return f
}

// seedLot inserts a lot at a deterministic created_at so FIFO order is stable.
func seedLot(t *testing.T, pool *pgxpool.Pool, f fixtures, lotNumber string, qty, unitCost float64, createdAt time.Time) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO lots (lot_number, product_id, warehouse_id, quantity, initial_quantity, unit_cost, created_at, received_at)
		 VALUES ($1, $2, $3, $4, $4, $5, $6, $6)
		 RETURNING id`,
		lotNumber, f.productID, f.warehouseID, qty, unitCost, createdAt,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed lot %s: %v", lotNumber, err)
	}
	return id
}

func newService(pool *pgxpool.Pool) *sale.Service {
	return sale.NewService(sale.NewRepository(pool), lot.NewRepository(pool), pool)
}

func lotQty(t *testing.T, pool *pgxpool.Pool, id string) float64 {
	t.Helper()
	var q float64
	err := pool.QueryRow(context.Background(),
		`SELECT quantity FROM lots WHERE id = $1`, id,
	).Scan(&q)
	if err != nil {
		t.Fatalf("read lot qty: %v", err)
	}
	return q
}

func TestCreateSale_FIFO_AllocatesOldestLotsFirst(t *testing.T) {
	pool := testutil.SetupDB(t)
	svc := newService(pool)
	f := seedFixtures(t, pool)

	now := time.Now()
	oldest := seedLot(t, pool, f, "L1", 5, 10, now.Add(-3*time.Hour))
	middle := seedLot(t, pool, f, "L2", 5, 12, now.Add(-2*time.Hour))
	newest := seedLot(t, pool, f, "L3", 5, 14, now.Add(-1*time.Hour))

	// Sell 7 — should drain L1 fully (5) and 2 from L2; L3 untouched.
	s, err := svc.Create(context.Background(), sale.CreateSaleRequest{
		ProductID:   f.productID,
		WarehouseID: f.warehouseID,
		BuyerName:   "alice",
		Qty:         7,
		SellPrice:   20,
	})
	if err != nil {
		t.Fatalf("create sale: %v", err)
	}
	if len(s.Allocations) != 2 {
		t.Fatalf("expected 2 allocations, got %d", len(s.Allocations))
	}
	if s.Allocations[0].LotID != oldest || s.Allocations[0].Qty != 5 {
		t.Errorf("first allocation should be 5 from oldest lot, got lot=%s qty=%v", s.Allocations[0].LotID, s.Allocations[0].Qty)
	}
	if s.Allocations[1].LotID != middle || s.Allocations[1].Qty != 2 {
		t.Errorf("second allocation should be 2 from middle lot, got lot=%s qty=%v", s.Allocations[1].LotID, s.Allocations[1].Qty)
	}
	if got := lotQty(t, pool, oldest); got != 0 {
		t.Errorf("oldest lot quantity: want 0, got %v", got)
	}
	if got := lotQty(t, pool, middle); got != 3 {
		t.Errorf("middle lot quantity: want 3, got %v", got)
	}
	if got := lotQty(t, pool, newest); got != 5 {
		t.Errorf("newest lot quantity: want 5, got %v", got)
	}
}

func TestCreateSale_InsufficientStock(t *testing.T) {
	pool := testutil.SetupDB(t)
	svc := newService(pool)
	f := seedFixtures(t, pool)
	seedLot(t, pool, f, "L1", 3, 10, time.Now())

	_, err := svc.Create(context.Background(), sale.CreateSaleRequest{
		ProductID: f.productID, WarehouseID: f.warehouseID, BuyerName: "alice", Qty: 5, SellPrice: 20,
	})
	if !errors.Is(err, sale.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

// TestCreateSale_Concurrent_NoOverAllocation drives N goroutines requesting
// the same product/warehouse with combined demand exceeding supply. The
// invariants checked here are exactly the ones the FOR UPDATE lock protects:
//
//   - total allocated never exceeds total seeded
//   - no lot row ends up negative
//   - sum(allocations) per lot ≤ initial qty of that lot
//   - successful sales account for exactly the missing stock
func TestCreateSale_Concurrent_NoOverAllocation(t *testing.T) {
	pool := testutil.SetupDB(t)
	svc := newService(pool)
	f := seedFixtures(t, pool)

	const (
		lotsCount       = 4
		qtyPerLot       = 5.0
		totalSeeded     = lotsCount * qtyPerLot // 20
		concurrentSales = 10
		qtyPerSale      = 3.0 // demand 30, supply 20 → 6 should succeed
	)

	now := time.Now()
	for i := 0; i < lotsCount; i++ {
		seedLot(t, pool, f, lotNumberFor(i), qtyPerLot, 10, now.Add(-time.Duration(lotsCount-i)*time.Hour))
	}

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		successes   int
		insufficient int
		otherErrs   []error
	)
	for i := 0; i < concurrentSales; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Create(context.Background(), sale.CreateSaleRequest{
				ProductID: f.productID, WarehouseID: f.warehouseID,
				BuyerName: "buyer", Qty: qtyPerSale, SellPrice: 20,
			})
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				successes++
			case errors.Is(err, sale.ErrInsufficientStock):
				insufficient++
			default:
				otherErrs = append(otherErrs, err)
			}
		}()
	}
	wg.Wait()

	if len(otherErrs) > 0 {
		t.Fatalf("unexpected errors: %v", otherErrs)
	}

	expectedSuccess := int(math.Floor(totalSeeded / qtyPerSale)) // 6
	if successes != expectedSuccess {
		t.Errorf("successful sales: want %d, got %d (insufficient=%d)", expectedSuccess, successes, insufficient)
	}

	// Aggregate invariants.
	var totalRemainingLot, totalAllocated, minLotQty float64
	err := pool.QueryRow(context.Background(),
		`SELECT COALESCE(SUM(quantity), 0), COALESCE(MIN(quantity), 0) FROM lots WHERE product_id = $1`, f.productID,
	).Scan(&totalRemainingLot, &minLotQty)
	if err != nil {
		t.Fatalf("aggregate lots: %v", err)
	}
	err = pool.QueryRow(context.Background(),
		`SELECT COALESCE(SUM(qty), 0) FROM sale_allocations sa
		 JOIN sales s ON s.id = sa.sale_id
		 WHERE s.product_id = $1`, f.productID,
	).Scan(&totalAllocated)
	if err != nil {
		t.Fatalf("aggregate allocations: %v", err)
	}

	if minLotQty < 0 {
		t.Errorf("invariant violated: a lot has negative quantity (%v)", minLotQty)
	}
	if totalRemainingLot+totalAllocated != totalSeeded {
		t.Errorf("conservation broken: remaining %v + allocated %v != seeded %v",
			totalRemainingLot, totalAllocated, totalSeeded)
	}

	// Per-lot: allocations against a lot never exceed its initial quantity.
	rows, err := pool.Query(context.Background(),
		`SELECT l.id, l.initial_quantity, COALESCE(SUM(sa.qty), 0)
		 FROM lots l
		 LEFT JOIN sale_allocations sa ON sa.lot_id = l.id
		 WHERE l.product_id = $1
		 GROUP BY l.id, l.initial_quantity`, f.productID,
	)
	if err != nil {
		t.Fatalf("per-lot aggregate: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var initial, allocated float64
		if err := rows.Scan(&id, &initial, &allocated); err != nil {
			t.Fatalf("scan per-lot: %v", err)
		}
		if allocated > initial {
			t.Errorf("over-allocation on lot %s: allocated %v > initial %v", id, allocated, initial)
		}
	}
}

// TestAddPayment_Concurrent_NoOverpay verifies that concurrent partial
// payments cannot push paid_amount past the sale's total. With the buggy
// read-then-write version, all goroutines could observe an outdated
// paid_amount and all succeed, summing to more than the total.
func TestAddPayment_Concurrent_NoOverpay(t *testing.T) {
	pool := testutil.SetupDB(t)
	svc := newService(pool)
	f := seedFixtures(t, pool)
	seedLot(t, pool, f, "L1", 100, 5, time.Now().Add(-time.Hour))

	// Create a sale of 1 unit at 100 — total owed = 100.
	s, err := svc.Create(context.Background(), sale.CreateSaleRequest{
		ProductID: f.productID, WarehouseID: f.warehouseID, BuyerName: "alice", Qty: 1, SellPrice: 100,
	})
	if err != nil {
		t.Fatalf("create sale: %v", err)
	}

	const (
		goroutines      = 8
		amountPerCall   = 30.0
		saleTotal       = 100.0
		expectedSuccess = 3 // floor(100/30) = 3
	)
	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		successes   int
		overpayErrs int
		otherErrs   []error
	)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.AddPayment(context.Background(), s.ID, amountPerCall)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				successes++
			case errors.Is(err, sale.ErrPaymentExceedsTotal):
				overpayErrs++
			default:
				otherErrs = append(otherErrs, err)
			}
		}()
	}
	wg.Wait()

	if len(otherErrs) > 0 {
		t.Fatalf("unexpected errors: %v", otherErrs)
	}
	if successes != expectedSuccess {
		t.Errorf("successful payments: want %d, got %d (overpay=%d)", expectedSuccess, successes, overpayErrs)
	}

	var paid float64
	err = pool.QueryRow(context.Background(),
		`SELECT paid_amount FROM sales WHERE id = $1`, s.ID,
	).Scan(&paid)
	if err != nil {
		t.Fatalf("read paid: %v", err)
	}
	if paid > saleTotal {
		t.Errorf("paid_amount %v exceeds sale total %v", paid, saleTotal)
	}
	if paid != float64(expectedSuccess)*amountPerCall {
		t.Errorf("paid_amount: want %v, got %v", float64(expectedSuccess)*amountPerCall, paid)
	}
}

func TestCancel_RestoresLotQuantity(t *testing.T) {
	pool := testutil.SetupDB(t)
	svc := newService(pool)
	f := seedFixtures(t, pool)
	lotID := seedLot(t, pool, f, "L1", 10, 5, time.Now().Add(-time.Hour))

	s, err := svc.Create(context.Background(), sale.CreateSaleRequest{
		ProductID: f.productID, WarehouseID: f.warehouseID, BuyerName: "alice", Qty: 7, SellPrice: 20,
	})
	if err != nil {
		t.Fatalf("create sale: %v", err)
	}
	if got := lotQty(t, pool, lotID); got != 3 {
		t.Fatalf("after sale: want lot qty 3, got %v", got)
	}

	if err := svc.Cancel(context.Background(), s.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	if got := lotQty(t, pool, lotID); got != 10 {
		t.Errorf("after cancel: want lot qty 10, got %v", got)
	}

	var status string
	err = pool.QueryRow(context.Background(),
		`SELECT shipment_status FROM sales WHERE id = $1`, s.ID,
	).Scan(&status)
	if err != nil {
		t.Fatalf("read sale status: %v", err)
	}
	if status != "canceled" {
		t.Errorf("sale shipment_status: want canceled, got %s", status)
	}

	var allocCount int
	err = pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM sale_allocations WHERE sale_id = $1`, s.ID,
	).Scan(&allocCount)
	if err != nil {
		t.Fatalf("count allocations: %v", err)
	}
	if allocCount != 0 {
		t.Errorf("after cancel: want 0 allocations, got %d", allocCount)
	}
}

func lotNumberFor(i int) string {
	return "LOT-" + string(rune('A'+i))
}
