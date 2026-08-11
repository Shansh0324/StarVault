package repository

import (
	"context"
	"database/sql"
	"time"
)

type AppRepository struct {
	DB *sql.DB
}

type AppRecord struct {
	ID         string
	Name       string
	SecretHash string
	CreatedAt  time.Time
}

func (r *AppRepository) CreateApp(ctx context.Context, name, secretHash string) (string, error) {
	var id string
	err := r.DB.QueryRowContext(ctx, `
		INSERT INTO apps (name, secret_hash)
		VALUES ($1, $2)
		RETURNING id
	`, name, secretHash).Scan(&id)
	
	return id, err
}

func (r *AppRepository) GetAppByID(ctx context.Context, id string) (*AppRecord, error) {
	record := &AppRecord{}
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, name, secret_hash, created_at
		FROM apps
		WHERE id = $1
	`, id).Scan(&record.ID, &record.Name, &record.SecretHash, &record.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	return record, nil
}
