package handlers

import (
	"encoding/json"
	"net/http"
	"starvault/core/internal/dtos"
	"starvault/core/internal/services"
)

type AuthHandler struct {
	AuthService *services.AuthService
}

func sendError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(dtos.AuthResponse{Error: msg})
}

// POST /internal/users
func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dtos.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	id, err := h.AuthService.Register(r.Context(), req)
	if err != nil {
		if err.Error() == "Email and password are required" {
			sendError(w, http.StatusBadRequest, err.Error())
		} else if err.Error() == "Email already exists" {
			sendError(w, http.StatusConflict, err.Error())
		} else {
			sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dtos.AuthResponse{UserID: id})
}

// POST /internal/users/verify
func (h *AuthHandler) VerifyUser(w http.ResponseWriter, r *http.Request) {
	var req dtos.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	id, err := h.AuthService.Login(r.Context(), req)
	if err != nil {
		if err.Error() == "Invalid email or password" {
			sendError(w, http.StatusUnauthorized, err.Error())
		} else {
			sendError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dtos.AuthResponse{UserID: id})
}
