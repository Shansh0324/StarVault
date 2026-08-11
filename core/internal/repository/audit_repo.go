package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type AuditLog struct {
	ID               string `json:"id"`
	UserID           string `json:"userId"`
	AppID            string `json:"appId"`
	Action           string `json:"action"`
	Scope            string `json:"scope"`
	EventHash        string `json:"eventHash"`
	Timestamp        string `json:"timestamp"`
	BlockchainStatus string `json:"blockchainStatus"`
}

type AuditRepository struct {
	DB *sql.DB
}

// LogEvent inserts a new PENDING audit event.
// Note: We insert "" for event_hash to satisfy NOT NULL if allowed,
// or a temporary UUID/placeholder if the schema rejects "".
func (r *AuditRepository) LogEvent(ctx context.Context, userID, appID, action, scope string) error {
	query := `
		INSERT INTO audit_logs (user_id, app_id, action, scope, event_hash, blockchain_status)
		VALUES ($1, $2, $3, $4, '', 'PENDING')
	`
	_, err := r.DB.ExecContext(ctx, query, userID, appID, action, scope)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}
	return nil
}

// FetchPending retrieves up to limit PENDING logs ordered by timestamp asc.
func (r *AuditRepository) FetchPending(ctx context.Context, limit int) ([]AuditLog, error) {
	query := `
		SELECT id, user_id, app_id, action, COALESCE(scope, ''), event_hash, timestamp, blockchain_status
		FROM audit_logs
		WHERE blockchain_status = 'PENDING'
		ORDER BY timestamp ASC
		LIMIT $1
	`
	rows, err := r.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending audits: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.AppID, &log.Action, &log.Scope, &log.EventHash, &log.Timestamp, &log.BlockchainStatus); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// GetLastHash gets the event_hash of the most recently COMMITTED log.
func (r *AuditRepository) GetLastHash(ctx context.Context) (string, error) {
	query := `
		SELECT event_hash
		FROM audit_logs
		WHERE blockchain_status = 'COMMITTED'
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var hash string
	err := r.DB.QueryRowContext(ctx, query).Scan(&hash)
	if err == sql.ErrNoRows {
		// Genesis block equivalent: return standard genesis hash
		return "0000000000000000000000000000000000000000000000000000000000000000", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get last hash: %w", err)
	}
	return hash, nil
}

// Commit updates the event_hash and marks as COMMITTED.
func (r *AuditRepository) Commit(ctx context.Context, id, hash string) error {
	query := `
		UPDATE audit_logs
		SET event_hash = $1, blockchain_status = 'COMMITTED'
		WHERE id = $2
	`
	_, err := r.DB.ExecContext(ctx, query, hash, id)
	if err != nil {
		return fmt.Errorf("failed to commit audit log: %w", err)
	}
	return nil
}
