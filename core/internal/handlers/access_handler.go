package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"starvault/core/internal/dtos"
	"starvault/core/internal/services"
)

type AccessHandler struct {
	AccessService *services.AccessService
}

func (h *AccessHandler) AccessData(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"missing X-User-ID header"}`, http.StatusUnauthorized)
		return
	}

	var req dtos.AccessDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	res, err := h.AccessService.AccessData(userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "forbidden") {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
