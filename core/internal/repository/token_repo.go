package repository

import (
	"context"
	"database/sql"
	"time"
)

type TokenRepository struct {
	DB *sql.DB
}

type TokenRecord struct {
	ID        string
	UserID    string
	AppID     string
	TokenHash string
	Scopes    string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (r *TokenRepository) CreateToken(ctx context.Context, userID, appID, tokenHash, scopes string, expiresAt time.Time) (string, error) {
	var id string
	err := r.DB.QueryRowContext(ctx, `
		INSERT INTO access_tokens (user_id, app_id, token_hash, scopes, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, appID, tokenHash, scopes, expiresAt).Scan(&id)
	
	return id, err
}

func (r *TokenRepository) GetTokenByHash(ctx context.Context, tokenHash string) (*TokenRecord, error) {
	record := &TokenRecord{}
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, user_id, app_id, token_hash, scopes, expires_at, revoked_at, created_at
		FROM access_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(&record.ID, &record.UserID, &record.AppID, &record.TokenHash, &record.Scopes, &record.ExpiresAt, &record.RevokedAt, &record.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *TokenRepository) RevokeToken(ctx context.Context, tokenHash, userID string) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE access_tokens
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE token_hash = $1 AND user_id = $2 AND revoked_at IS NULL
	`, tokenHash, userID)
	if err != nil {
		return err
	}
	
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows // Either doesn't exist, already revoked, or doesn't belong to user
	}
	
	return nil
}
