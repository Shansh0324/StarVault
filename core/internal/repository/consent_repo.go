package repository

import (
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
}

func (r *ConsentRepository) CreateConsent(userID, appID, scopesJSON, purpose string, expiresAt time.Time) (string, time.Time, error) {
	var id string
	var createdAt time.Time

	err := r.DB.QueryRow(`
		INSERT INTO consents (user_id, app_id, scopes, purpose, status, expires_at)
		VALUES ($1, $2, $3, $4, 'ACTIVE', $5)
		RETURNING id, created_at
	`, userID, appID, scopesJSON, purpose, expiresAt).Scan(&id, &createdAt)

	return id, createdAt, err
}

func (r *ConsentRepository) GetConsentByIDAndUserID(id, userID string) (*ConsentRecord, error) {
	record := &ConsentRecord{}
	err := r.DB.QueryRow(`
		SELECT id, user_id, app_id, scopes, purpose, status, expires_at, created_at, revoked_at
		FROM consents
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(
		&record.ID, &record.UserID, &record.AppID, &record.Scopes,
		&record.Purpose, &record.Status, &record.ExpiresAt,
		&record.CreatedAt, &record.RevokedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("consent not found")
		}
		return nil, err
	}
	return record, nil
}

func (r *ConsentRepository) ListConsentsByUserID(userID string) ([]*ConsentRecord, error) {
	rows, err := r.DB.Query(`
		SELECT id, user_id, app_id, scopes, purpose, status, expires_at, created_at, revoked_at
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
			&record.CreatedAt, &record.RevokedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (r *ConsentRepository) RevokeConsent(id, userID string) error {
	res, err := r.DB.Exec(`
		UPDATE consents
		SET status = 'REVOKED', revoked_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND status = 'ACTIVE'
	`, id, userID)
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("consent not found or already revoked")
	}

	return nil
}

func (r *ConsentRepository) GetActiveConsentForApp(userID, appID string) (*ConsentRecord, error) {
	record := &ConsentRecord{}
	err := r.DB.QueryRow(`
		SELECT id, user_id, app_id, scopes, purpose, status, expires_at, created_at, revoked_at
		FROM consents
		WHERE user_id = $1 AND app_id = $2 AND status = 'ACTIVE'
		ORDER BY created_at DESC LIMIT 1
	`, userID, appID).Scan(
		&record.ID, &record.UserID, &record.AppID, &record.Scopes,
		&record.Purpose, &record.Status, &record.ExpiresAt,
		&record.CreatedAt, &record.RevokedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("consent not found")
		}
		return nil, err
	}
	return record, nil
}
