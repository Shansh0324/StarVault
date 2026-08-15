package repository

import (
	"context"
	"database/sql"
	"errors"
)

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (string, error) {
	var id string
	err := r.DB.QueryRowContext(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", email, passwordHash).Scan(&id)
	return id, err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (string, string, error) {
	var dbID, dbHash string
	err := r.DB.QueryRowContext(ctx, "SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&dbID, &dbHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", errors.New("user not found")
		}
		return "", "", err
	}
	return dbID, dbHash, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (string, string, error) {
	var email, did sql.NullString
	err := r.DB.QueryRowContext(ctx, "SELECT email, did FROM users WHERE id = $1", id).Scan(&email, &did)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", errors.New("user not found")
		}
		return "", "", err
	}
	return email.String, did.String, nil
}

func (r *UserRepository) SetDID(ctx context.Context, id, did string) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE users SET did = $1 WHERE id = $2", did, id)
	return err
}
