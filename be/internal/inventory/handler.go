package inventory

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lifosmin/admin-backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RecordMovement(w http.ResponseWriter, r *http.Request) {
	var req CreateMovementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.LotID == "" || req.MovementType == "" || req.Quantity == 0 {
		http.Error(w, `{"error":"lot_id, movement_type, and quantity required"}`, http.StatusBadRequest)
		return
	}

	performedBy := middleware.UserIDFromCtx(r.Context())
	m, err := h.service.RecordMovement(r.Context(), req, performedBy)
	if err != nil {
		if err == ErrInsufficientStock {
			http.Error(w, `{"error":"insufficient stock"}`, http.StatusConflict)
			return
		}
		if err == ErrInvalidMovement {
			http.Error(w, `{"error":"invalid movement type"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) ListByLot(w http.ResponseWriter, r *http.Request) {
	lotID := chi.URLParam(r, "id")
	movements, err := h.service.ListByLot(r.Context(), lotID)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movements)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	movements, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movements)
}
