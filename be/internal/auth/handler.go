package auth

import (
	"encoding/json"
	"net/http"

	"github.com/lifosmin/admin-backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password required"}`, http.StatusBadRequest)
		return
	}

	tokenPair, cookies, err := h.service.Login(r.Context(), req)
	if err != nil {
		if err == ErrInvalidCredentials {
			http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	for _, c := range cookies {
		http.SetCookie(w, c)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "logged in",
		"access_token": tokenPair.AccessToken,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, `{"error":"no refresh token found"}`, http.StatusUnauthorized)
		return
	}

	_, cookies, err := h.service.Refresh(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, `{"error":"unable to refresh"}`, http.StatusUnauthorized)
		return
	}

	for _, c := range cookies {
		http.SetCookie(w, c)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "refreshed"})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookies := h.service.Logout()
	for _, c := range cookies {
		http.SetCookie(w, c)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	role := middleware.RoleFromCtx(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":   userID,
		"role": role,
	})
}
