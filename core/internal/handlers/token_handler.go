package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"starvault/core/internal/dtos"
	"starvault/core/internal/services"
)

type TokenHandler struct {
	TokenService *services.TokenService
}

func (h *TokenHandler) IssueToken(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"missing X-User-ID header"}`, http.StatusUnauthorized)
		return
	}

	var req dtos.TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	rawToken, err := h.TokenService.GenerateToken(r.Context(), userID, req.AppID, req.AppSecret, req.Scope)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "forbidden") {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dtos.TokenResponse{
		Token:     rawToken,
		ExpiresIn: 900, // 15 mins
	})
}

func (h *TokenHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"missing X-User-ID header"}`, http.StatusUnauthorized)
		return
	}

	var req dtos.RevokeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	err := h.TokenService.RevokeToken(r.Context(), userID, req.Token)
	if err != nil {
		// Generic error to prevent leaking token existence state to non-owners
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
