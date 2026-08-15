package handlers

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"starvault/core/internal/blockchain"
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

// VerifyAudit returns the audit event, its Merkle proof, and on-chain tx hash for independent verification.
func (h *AuditHandler) VerifyAudit(w http.ResponseWriter, r *http.Request) {
	eventID := r.URL.Query().Get("eventId")
	if eventID == "" {
		http.Error(w, `{"error":"eventId query param required"}`, http.StatusBadRequest)
		return
	}

	// Get the audit event
	var auditLog repository.AuditLog
	err := h.AuditRepo.DB.QueryRow(
		`SELECT id, user_id, app_id, action, COALESCE(scope, ''), event_hash, timestamp, blockchain_status
		 FROM audit_logs WHERE id = $1`, eventID,
	).Scan(&auditLog.ID, &auditLog.UserID, &auditLog.AppID, &auditLog.Action, &auditLog.Scope, &auditLog.EventHash, &auditLog.Timestamp, &auditLog.BlockchainStatus)
	if err != nil {
		http.Error(w, `{"error":"audit event not found"}`, http.StatusNotFound)
		return
	}

	// If not yet batched, return event only
	if auditLog.BlockchainStatus != "COMMITTED" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"event":   auditLog,
			"status":  "PENDING_BATCH",
			"message": "Event recorded but not yet anchored on blockchain. Awaiting next batch cycle.",
		})
		return
	}

	// Get the batch
	batch, err := h.AuditRepo.GetBatchByEventID(r.Context(), eventID)
	if err != nil {
		http.Error(w, `{"error":"batch not found for event"}`, http.StatusNotFound)
		return
	}

	// Get all events in the batch to reconstruct the Merkle tree
	batchEvents, err := h.AuditRepo.GetBatchEventHashes(r.Context(), batch.ID)
	if err != nil {
		http.Error(w, `{"error":"failed to reconstruct batch"}`, http.StatusInternalServerError)
		return
	}

	// Build Merkle tree and generate proof
	leaves := make([][]byte, len(batchEvents))
	targetIndex := -1
	for i, e := range batchEvents {
		hashBytes, _ := hex.DecodeString(e.EventHash)
		leaves[i] = hashBytes
		if e.ID == eventID {
			targetIndex = i
		}
	}

	if targetIndex == -1 {
		http.Error(w, `{"error":"event not found in batch"}`, http.StatusInternalServerError)
		return
	}

	tree := &blockchain.MerkleTree{Leaves: leaves}
	proof := tree.Proof(targetIndex)

	// Encode proof steps as hex for JSON
	type hexProofStep struct {
		Hash    string `json:"hash"`
		IsRight bool   `json:"isRight"`
	}
	hexProof := make([]hexProofStep, len(proof))
	for i, p := range proof {
		hexProof[i] = hexProofStep{Hash: hex.EncodeToString(p.Hash), IsRight: p.IsRight}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event":       auditLog,
		"batch":       batch,
		"merkleRoot":  batch.MerkleRoot,
		"merkleProof": hexProof,
		"leafIndex":   targetIndex,
		"verified":    blockchain.VerifyProof(leaves[targetIndex], tree.Root(), proof),
	})
}
