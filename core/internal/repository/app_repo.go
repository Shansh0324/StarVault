package repository

import (
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

func (r *AppRepository) CreateApp(name, secretHash string) (string, error) {
	var id string
	err := r.DB.QueryRow(`
		INSERT INTO apps (name, secret_hash)
		VALUES ($1, $2)
		RETURNING id
	`, name, secretHash).Scan(&id)
	
	return id, err
}

func (r *AppRepository) GetAppByID(id string) (*AppRecord, error) {
	record := &AppRecord{}
	err := r.DB.QueryRow(`
		SELECT id, name, secret_hash, created_at
		FROM apps
		WHERE id = $1
	`, id).Scan(&record.ID, &record.Name, &record.SecretHash, &record.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	return record, nil
}
