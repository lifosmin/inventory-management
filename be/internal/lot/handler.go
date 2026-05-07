package lot

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	repo    *Repository
	logger  *slog.Logger
}

func NewHandler(service *Service, repo *Repository) *Handler {
	return &Handler{service: service, repo: repo, logger: slog.Default()}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateLotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.LotNumber == "" || req.ProductID == "" || req.WarehouseID == "" {
		http.Error(w, `{"error":"lot_number, product_id, and warehouse_id required"}`, http.StatusBadRequest)
		return
	}
	if req.Quantity <= 0 || req.UnitCost < 0 {
		http.Error(w, `{"error":"quantity must be positive, unit_cost non-negative"}`, http.StatusBadRequest)
		return
	}

	l, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("lot create failed", "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(l)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	l, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == ErrLotNotFound {
			http.Error(w, `{"error":"lot not found"}`, http.StatusNotFound)
			return
		}
		h.logger.Error("lot getbyid failed", "id", id, "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(l)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	params := ListParams{
		SortBy:         r.URL.Query().Get("sort_by"),
		SortDir:        r.URL.Query().Get("sort_dir"),
		Search:         r.URL.Query().Get("search"),
		ProductID:      r.URL.Query().Get("product_id"),
		WarehouseID:    r.URL.Query().Get("warehouse_id"),
		ShipmentStatus: r.URL.Query().Get("shipment_status"),
		PaymentStatus:  r.URL.Query().Get("payment_status"),
		DateFrom:       r.URL.Query().Get("date_from"),
		DateTo:         r.URL.Query().Get("date_to"),
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		fmt.Sscan(v, &params.Limit)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		fmt.Sscan(v, &params.Offset)
	}

	result, err := h.service.List(r.Context(), params)
	if err != nil {
		h.logger.Error("lot list failed", "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req UpdateLotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	l, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		if err == ErrLotNotFound {
			http.Error(w, `{"error":"lot not found"}`, http.StatusNotFound)
			return
		}
		h.logger.Error("lot update failed", "id", id, "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(l)
}

func (h *Handler) UpdateShipmentStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		ShipmentStatus string `json:"shipment_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if body.ShipmentStatus == "" {
		http.Error(w, `{"error":"shipment_status required"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateShipmentStatus(r.Context(), id, body.ShipmentStatus); err != nil {
		h.logger.Error("lot update shipment status failed", "id", id, "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "updated"})
}

func (h *Handler) AddPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req AddPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		http.Error(w, `{"error":"amount must be positive"}`, http.StatusBadRequest)
		return
	}

	l, err := h.repo.AddPayment(r.Context(), id, req.Amount)
	if err != nil {
		if err == ErrPaymentExceedsTotal {
			http.Error(w, `{"error":"payment would exceed total owed"}`, http.StatusUnprocessableEntity)
			return
		}
		h.logger.Error("lot add payment failed", "id", id, "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	if l == nil {
		http.Error(w, `{"error":"lot not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(l)
}

func (h *Handler) MarkDelivered(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req DeliverLotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.DeliveredDate == "" {
		http.Error(w, `{"error":"delivered_date required"}`, http.StatusBadRequest)
		return
	}
	if req.AdditionalCost < 0 {
		http.Error(w, `{"error":"additional_cost cannot be negative"}`, http.StatusBadRequest)
		return
	}
	if req.ActualReceivedQty <= 0 {
		http.Error(w, `{"error":"actual_received_qty must be positive"}`, http.StatusBadRequest)
		return
	}

	l, err := h.repo.MarkDelivered(r.Context(), id, req.DeliveredDate, req.AdditionalCost, req.ActualReceivedQty)
	if err != nil {
		if err == ErrLotNotFound {
			http.Error(w, `{"error":"lot not found"}`, http.StatusNotFound)
			return
		}
		if err == ErrAlreadyDelivered {
			http.Error(w, `{"error":"lot is already delivered"}`, http.StatusConflict)
			return
		}
		if err == ErrQtyExceedsOrdered {
			http.Error(w, `{"error":"actual received qty exceeds originally ordered qty"}`, http.StatusUnprocessableEntity)
			return
		}
		if err == ErrActualQtyBelowSold {
			http.Error(w, `{"error":"actual received qty cannot be less than already sold qty"}`, http.StatusUnprocessableEntity)
			return
		}
		h.logger.Error("lot mark delivered failed", "id", id, "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(l)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Cancel(r.Context(), id); err != nil {
		switch err {
		case ErrLotNotFound:
			http.Error(w, `{"error":"lot not found"}`, http.StatusNotFound)
		case ErrAlreadyDelivered:
			http.Error(w, `{"error":"lot is already delivered and cannot be canceled"}`, http.StatusConflict)
		case ErrLotAlreadyCanceled:
			http.Error(w, `{"error":"lot is already canceled"}`, http.StatusConflict)
		case ErrLotHasDependentSales:
			http.Error(w, `{"error":"lot has active dependent sales and cannot be canceled"}`, http.StatusConflict)
		default:
			h.logger.Error("lot cancel failed", "id", id, "error", err)
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "canceled"})
}

func (h *Handler) GetAvailableQty(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		http.Error(w, `{"error":"product_id required"}`, http.StatusBadRequest)
		return
	}
	items, err := h.repo.GetAvailableQtyByProduct(r.Context(), productID)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []AvailableQtyItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
