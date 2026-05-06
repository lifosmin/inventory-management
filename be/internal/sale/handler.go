package sale

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.ProductID == "" || req.WarehouseID == "" || req.BuyerName == "" || req.Qty <= 0 || req.SellPrice < 0 {
		http.Error(w, `{"error":"product_id, warehouse_id, buyer_name, qty (>0), and sell_price required"}`, http.StatusBadRequest)
		return
	}

	s, err := h.service.Create(r.Context(), req)
	if err != nil {
		if err == ErrInsufficientStock {
			http.Error(w, `{"error":"insufficient stock for this sale"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sales, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sales)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateStatus(r.Context(), id, req); err != nil {
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

	s, err := h.service.AddPayment(r.Context(), id, req.Amount)
	if err != nil {
		if err == ErrPaymentExceedsTotal {
			http.Error(w, `{"error":"payment would exceed total owed"}`, http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	if s == nil {
		http.Error(w, `{"error":"sale not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	d, err := h.service.GetDashboard(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}
