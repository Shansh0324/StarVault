package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type ConsentRepository struct {
	DB *sql.DB
}

type ConsentRecord struct {
	ID        string
	UserID    string
	AppID     string
	Scopes    string // JSON encoded array
	Purpose   string
	Status    string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt sql.NullTime
	Policies  string // JSON encoded map
}

func (r *ConsentRepository) CreateConsent(ctx context.Context, userID, appID, scopesJSON, purpose, policiesJSON string, expiresAt time.Time) (string, time.Time, error) {
	var id string
	var createdAt time.Time

	err := r.DB.QueryRowContext(ctx, `
		INSERT INTO consents (user_id, app_id, scopes, purpose, status, expires_at, policies)
		VALUES ($1, $2, $3, $4, 'ACTIVE', $5, $6)
		RETURNING id, created_at
	`, userID, appID, scopesJSON, purpose, expiresAt, policiesJSON).Scan(&id, &createdAt)

	return id, createdAt, err
}

func (r *ConsentRepository) GetConsentByIDAndUserID(ctx context.Context, id, userID string) (*ConsentRecord, error) {
	record := &ConsentRecord{}
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, user_id, app_id, scopes, purpose, status, expires_at, created_at, revoked_at, policies
		FROM consents
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(
		&record.ID, &record.UserID, &record.AppID, &record.Scopes,
		&record.Purpose, &record.Status, &record.ExpiresAt,
		&record.CreatedAt, &record.RevokedAt, &record.Policies,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("consent not found")
		}
		return nil, err
	}
	return record, nil
}

func (r *ConsentRepository) ListConsentsByUserID(ctx context.Context, userID string) ([]*ConsentRecord, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, user_id, app_id, scopes, purpose, status, expires_at, created_at, revoked_at, policies
		FROM consents
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*ConsentRecord
	for rows.Next() {
		record := &ConsentRecord{}
		if err := rows.Scan(
			&record.ID, &record.UserID, &record.AppID, &record.Scopes,
			&record.Purpose, &record.Status, &record.ExpiresAt,
			&record.CreatedAt, &record.RevokedAt, &record.Policies,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (r *ConsentRepository) RevokeConsent(ctx context.Context, id, userID string) (string, error) {
	var appID string
	err := r.DB.QueryRowContext(ctx, `
		UPDATE consents
		SET status = 'REVOKED', revoked_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND status = 'ACTIVE'
		RETURNING app_id
	`, id, userID).Scan(&appID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("consent not found or already revoked")
		}
		return "", err
	}

	return appID, nil
}

func (r *ConsentRepository) GetActiveConsentForApp(ctx context.Context, userID, appID string) (*ConsentRecord, error) {
	record := &ConsentRecord{}
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, user_id, app_id, scopes, purpose, status, expires_at, created_at, revoked_at, policies
		FROM consents
		WHERE user_id = $1 AND app_id = $2 AND status = 'ACTIVE'
		ORDER BY created_at DESC LIMIT 1
	`, userID, appID).Scan(
		&record.ID, &record.UserID, &record.AppID, &record.Scopes,
		&record.Purpose, &record.Status, &record.ExpiresAt,
		&record.CreatedAt, &record.RevokedAt, &record.Policies,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("consent not found")
		}
		return nil, err
	}
	return record, nil
}
