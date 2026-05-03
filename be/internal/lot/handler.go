package lot

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	repo    *Repository
}

func NewHandler(service *Service, repo *Repository) *Handler {
	return &Handler{service: service, repo: repo}
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
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(l)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	lots, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lots)
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
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "updated"})
}
