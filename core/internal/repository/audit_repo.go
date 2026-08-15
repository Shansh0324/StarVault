package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type AuditLog struct {
	ID               string  `json:"id"`
	UserID           string  `json:"userId"`
	AppID            string  `json:"appId"`
	Action           string  `json:"action"`
	Scope            string  `json:"scope"`
	EventHash        string  `json:"eventHash"`
	Timestamp        string  `json:"timestamp"`
	BlockchainStatus string  `json:"blockchainStatus"`
	BatchID          *string `json:"batchId,omitempty"`
}

type AuditBatch struct {
	ID         string  `json:"id"`
	MerkleRoot string  `json:"merkleRoot"`
	TxHash     *string `json:"txHash,omitempty"`
	EventCount int     `json:"eventCount"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"createdAt"`
	AnchoredAt *string `json:"anchoredAt,omitempty"`
}

type AuditRepository struct {
	DB *sql.DB
}

// InsertPendingBatch inserts an audit event with PENDING_BATCH status (no blockchain yet).
func (r *AuditRepository) InsertPendingBatch(ctx context.Context, id, userID, appID, action, scope, hash string) error {
	query := `
		INSERT INTO audit_logs (id, user_id, app_id, action, scope, event_hash, blockchain_status)
		VALUES ($1, $2, $3, $4, $5, $6, 'PENDING_BATCH')
	`
	_, err := r.DB.ExecContext(ctx, query, id, userID, appID, action, scope, hash)
	if err != nil {
		return fmt.Errorf("failed to insert pending batch audit log: %w", err)
	}
	return nil
}

// FetchPendingBatch retrieves up to limit PENDING_BATCH logs for Merkle batching.
func (r *AuditRepository) FetchPendingBatch(ctx context.Context, limit int) ([]AuditLog, error) {
	query := `
		SELECT id, user_id, app_id, action, COALESCE(scope, ''), event_hash, timestamp, blockchain_status
		FROM audit_logs
		WHERE blockchain_status = 'PENDING_BATCH'
		ORDER BY timestamp ASC
		LIMIT $1
	`
	rows, err := r.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending batch audits: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.AppID, &l.Action, &l.Scope, &l.EventHash, &l.Timestamp, &l.BlockchainStatus); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// CreateBatchAndCommit creates a batch record and updates all referenced audit logs atomically.
func (r *AuditRepository) CreateBatchAndCommit(ctx context.Context, merkleRoot, txHash string, eventIDs []string) (string, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// Insert batch record
	var batchID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO audit_batches (merkle_root, tx_hash, event_count, status, anchored_at)
		 VALUES ($1, $2, $3, 'ANCHORED', NOW())
		 RETURNING id`,
		merkleRoot, txHash, len(eventIDs),
	).Scan(&batchID)
	if err != nil {
		return "", fmt.Errorf("failed to insert batch: %w", err)
	}

	// Update all audit logs in this batch
	_, err = tx.ExecContext(ctx,
		`UPDATE audit_logs
		 SET blockchain_status = 'COMMITTED', batch_id = $1
		 WHERE id = ANY($2)`,
		batchID, pq.Array(eventIDs),
	)
	if err != nil {
		return "", fmt.Errorf("failed to update audit logs for batch: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit batch tx: %w", err)
	}
	return batchID, nil
}

// GetBatchByEventID retrieves the batch associated with a specific audit event.
func (r *AuditRepository) GetBatchByEventID(ctx context.Context, eventID string) (*AuditBatch, error) {
	query := `
		SELECT b.id, b.merkle_root, b.tx_hash, b.event_count, b.status, b.created_at, b.anchored_at
		FROM audit_batches b
		INNER JOIN audit_logs a ON a.batch_id = b.id
		WHERE a.id = $1
	`
	var batch AuditBatch
	err := r.DB.QueryRowContext(ctx, query, eventID).Scan(
		&batch.ID, &batch.MerkleRoot, &batch.TxHash, &batch.EventCount,
		&batch.Status, &batch.CreatedAt, &batch.AnchoredAt,
	)
	if err != nil {
		return nil, fmt.Errorf("batch not found for event %s: %w", eventID, err)
	}
	return &batch, nil
}

// GetBatchEventHashes returns all event hashes in a batch, ordered by timestamp.
// Used to reconstruct the Merkle tree for proof generation.
func (r *AuditRepository) GetBatchEventHashes(ctx context.Context, batchID string) ([]AuditLog, error) {
	query := `
		SELECT id, user_id, app_id, action, COALESCE(scope, ''), event_hash, timestamp, blockchain_status
		FROM audit_logs
		WHERE batch_id = $1
		ORDER BY timestamp ASC
	`
	rows, err := r.DB.QueryContext(ctx, query, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch batch event hashes: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.AppID, &l.Action, &l.Scope, &l.EventHash, &l.Timestamp, &l.BlockchainStatus); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// GetLastHash gets the event_hash of the most recently inserted log (any status).
func (r *AuditRepository) GetLastHash(ctx context.Context) (string, error) {
	query := `
		SELECT event_hash
		FROM audit_logs
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var hash string
	err := r.DB.QueryRowContext(ctx, query).Scan(&hash)
	if err == sql.ErrNoRows {
		return "0000000000000000000000000000000000000000000000000000000000000000", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get last hash: %w", err)
	}
	return hash, nil
}

// LogEvent inserts a new PENDING audit event (fallback for no-NATS mode).
func (r *AuditRepository) LogEvent(ctx context.Context, userID, appID, action, scope string) error {
	query := `
		INSERT INTO audit_logs (user_id, app_id, action, scope, event_hash, blockchain_status)
		VALUES ($1, $2, $3, $4, '', 'PENDING_BATCH')
	`
	_, err := r.DB.ExecContext(ctx, query, userID, appID, action, scope)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}
	return nil
}
