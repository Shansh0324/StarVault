package handlers

import (
	"encoding/json"
	"net/http"
	"starvault/core/internal/repository"
)

type AuditHandler struct {
	AuditRepo *repository.AuditRepository
}

func (h *AuditHandler) GetLatestAudit(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, user_id, app_id, action, COALESCE(scope, ''), event_hash, timestamp, blockchain_status
		FROM audit_logs
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var log repository.AuditLog
	err := h.AuditRepo.DB.QueryRow(query).Scan(&log.ID, &log.UserID, &log.AppID, &log.Action, &log.Scope, &log.EventHash, &log.Timestamp, &log.BlockchainStatus)
	if err != nil {
		http.Error(w, "Audit not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}
