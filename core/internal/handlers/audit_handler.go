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
	appID := r.URL.Query().Get("appId")
	var query string
	var args []interface{}

	if appID != "" {
		query = `
			SELECT id, user_id, app_id, action, COALESCE(scope, ''), event_hash, timestamp, blockchain_status
			FROM audit_logs
			WHERE app_id = $1
			ORDER BY timestamp DESC
			LIMIT 1
		`
		args = append(args, appID)
	} else {
		query = `
			SELECT id, user_id, app_id, action, COALESCE(scope, ''), event_hash, timestamp, blockchain_status
			FROM audit_logs
			ORDER BY timestamp DESC
			LIMIT 1
		`
	}

	var log repository.AuditLog
	err := h.AuditRepo.DB.QueryRow(query, args...).Scan(&log.ID, &log.UserID, &log.AppID, &log.Action, &log.Scope, &log.EventHash, &log.Timestamp, &log.BlockchainStatus)
	if err != nil {
		http.Error(w, "Audit not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}
