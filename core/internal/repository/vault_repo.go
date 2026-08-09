package repository

import (
	"database/sql"
	"errors"
	"time"
)

type VaultRepository struct {
	DB *sql.DB
}

type VaultRecord struct {
	ID               string
	UserID           string
	DataType         string
	EncryptedPayload string
	CreatedAt        time.Time
}

func (r *VaultRepository) Create(userID, dataType, encryptedPayload string) (string, time.Time, error) {
	var id string
	var createdAt time.Time
	err := r.DB.QueryRow(`
		INSERT INTO vault_data (user_id, data_type, encrypted_payload)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, userID, dataType, encryptedPayload).Scan(&id, &createdAt)
	
	return id, createdAt, err
}

func (r *VaultRepository) GetByIDAndUserID(id, userID string) (*VaultRecord, error) {
	record := &VaultRecord{}
	err := r.DB.QueryRow(`
		SELECT id, user_id, data_type, encrypted_payload, created_at
		FROM vault_data
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&record.ID, &record.UserID, &record.DataType, &record.EncryptedPayload, &record.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("record not found")
		}
		return nil, err
	}
	return record, nil
}
