package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Valuation(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.InventoryValuation(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) Aging(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.LotAging(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		http.Error(w, `{"error":"from and to date required (YYYY-MM-DD)"}`, http.StatusBadRequest)
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		http.Error(w, `{"error":"invalid from date format"}`, http.StatusBadRequest)
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		http.Error(w, `{"error":"invalid to date format"}`, http.StatusBadRequest)
		return
	}

	report, err := h.service.Export(r.Context(), from, to)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "csv" {
		h.exportCSV(w, report)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *Handler) exportCSV(w http.ResponseWriter, rpt *ExportReport) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=report_%s_%s.csv", rpt.From, rpt.To))

	cw := csv.NewWriter(w)
	defer cw.Flush()

	cw.Write([]string{"INVENTORY REPORT", rpt.From, "to", rpt.To})
	cw.Write([]string{})

	cw.Write([]string{"SUMMARY"})
	cw.Write([]string{"Net Worth", fmt.Sprintf("%.0f", rpt.NetWorth)})
	cw.Write([]string{"Total Revenue", fmt.Sprintf("%.0f", rpt.TotalRevenue)})
	cw.Write([]string{"Total COGS", fmt.Sprintf("%.0f", rpt.TotalCOGS)})
	cw.Write([]string{"Profit/Loss", fmt.Sprintf("%.0f", rpt.ProfitLoss)})
	cw.Write([]string{})

	cw.Write([]string{"CURRENT STOCKS"})
	cw.Write([]string{"Product", "Warehouse", "Quantity", "Value"})
	for _, s := range rpt.Stocks {
		cw.Write([]string{s.ProductName, s.WarehouseName, fmt.Sprintf("%.0f", s.TotalQty), fmt.Sprintf("%.0f", s.TotalValue)})
	}
	cw.Write([]string{})

	cw.Write([]string{"RESTOCKS (in period)"})
	cw.Write([]string{"Lot Number", "Product", "Warehouse", "Qty", "Unit Cost", "Total Cost", "Supplier", "Date"})
	for _, r := range rpt.Restocks {
		cw.Write([]string{r.LotNumber, r.ProductName, r.WarehouseName,
			fmt.Sprintf("%.0f", r.Quantity), fmt.Sprintf("%.0f", r.UnitCost), fmt.Sprintf("%.0f", r.TotalCost),
			r.Supplier, r.CreatedAt.Format("2006-01-02")})
	}
	cw.Write([]string{})

	cw.Write([]string{"SALES (in period)"})
	cw.Write([]string{"Buyer", "Product", "Warehouse", "Qty", "Sell Price", "Total", "Date"})
	for _, s := range rpt.Sales {
		cw.Write([]string{s.BuyerName, s.ProductName, s.WarehouseName,
			fmt.Sprintf("%.0f", s.Qty), fmt.Sprintf("%.0f", s.SellPrice), fmt.Sprintf("%.0f", s.Total),
			s.CreatedAt.Format("2006-01-02")})
	}
}
