package handlers

import (
	"encoding/json"
	"net/http"
	"starvault/core/internal/services"
)

type WebAuthnHandler struct {
	WebAuthnService *services.WebAuthnService
}

func (h *WebAuthnHandler) RegisterBegin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, `{"error":"email required"}`, http.StatusBadRequest)
		return
	}

	options, sessionID, err := h.WebAuthnService.RegisterBegin(r.Context(), email)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"options":   options,
		"sessionId": sessionID,
	})
}

func (h *WebAuthnHandler) RegisterFinish(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	sessionID := r.URL.Query().Get("sessionId")

	err := h.WebAuthnService.RegisterFinish(r.Context(), email, sessionID, nil) // Mock response
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *WebAuthnHandler) LoginBegin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, `{"error":"email required"}`, http.StatusBadRequest)
		return
	}

	options, sessionID, err := h.WebAuthnService.LoginBegin(r.Context(), email)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"options":   options,
		"sessionId": sessionID,
	})
}

func (h *WebAuthnHandler) LoginFinish(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	sessionID := r.URL.Query().Get("sessionId")

	userID, err := h.WebAuthnService.LoginFinish(r.Context(), email, sessionID, nil)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"userId": userID,
	})
}
