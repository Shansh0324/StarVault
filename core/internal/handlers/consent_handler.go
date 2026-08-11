package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"starvault/core/internal/dtos"
	"starvault/core/internal/services"
)

type ConsentHandler struct {
	ConsentService *services.ConsentService
}

func (h *ConsentHandler) CreateConsent(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Missing X-User-ID header")
		return
	}

	var req dtos.CreateConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	res, err := h.ConsentService.CreateConsent(r.Context(), userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "unsupported") || strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "future") {
			sendError(w, http.StatusBadRequest, err.Error())
		} else if strings.Contains(err.Error(), "invalid app_id") {
			sendError(w, http.StatusNotFound, err.Error())
		} else {
			sendError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *ConsentHandler) GetConsent(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Missing X-User-ID header")
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	id := pathParts[len(pathParts)-1]

	res, err := h.ConsentService.GetConsent(r.Context(), id, userID)
	if err != nil {
		sendError(w, http.StatusNotFound, "Consent not found")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *ConsentHandler) ListConsents(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Missing X-User-ID header")
		return
	}

	res, err := h.ConsentService.ListConsents(r.Context(), userID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *ConsentHandler) RevokeConsent(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Missing X-User-ID header")
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	id := pathParts[len(pathParts)-2] // /internal/consents/{id}/revoke

	err := h.ConsentService.RevokeConsent(r.Context(), id, userID)
	if err != nil {
		sendError(w, http.StatusNotFound, "Consent not found or already revoked")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func (h *ConsentHandler) CheckConsent(w http.ResponseWriter, r *http.Request) {
	var req dtos.CheckConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	res, err := h.ConsentService.CheckConsent(r.Context(), req)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
