package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"starvault/core/internal/dtos"
	"starvault/core/internal/services"
)

type VaultHandler struct {
	VaultService *services.VaultService
}

func (h *VaultHandler) CreateVaultData(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Missing X-User-ID header")
		return
	}

	var req dtos.CreateVaultDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	res, err := h.VaultService.CreateVaultData(r.Context(), userID, req)
	if err != nil {
		if err.Error() == "dataType and data are required" || err.Error() == "data payload exceeds maximum allowed size" {
			sendError(w, http.StatusBadRequest, err.Error())
		} else {
			sendError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *VaultHandler) GetVaultData(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		sendError(w, http.StatusUnauthorized, "Missing X-User-ID header")
		return
	}

	// Extract ID from URL path: /internal/vault/data/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	id := pathParts[len(pathParts)-1]
	if id == "" || id == "data" {
		sendError(w, http.StatusBadRequest, "Missing vault ID")
		return
	}

	res, err := h.VaultService.GetVaultData(r.Context(), id, userID)
	if err != nil {
		if err.Error() == "record not found" {
			sendError(w, http.StatusNotFound, "Record not found")
		} else {
			sendError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
