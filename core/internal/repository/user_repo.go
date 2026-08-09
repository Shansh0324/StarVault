package repository

import (
	"database/sql"
	"errors"
)

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) CreateUser(email, passwordHash string) (string, error) {
	var id string
	err := r.DB.QueryRow("INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", email, passwordHash).Scan(&id)
	return id, err
}

func (r *UserRepository) GetUserByEmail(email string) (string, string, error) {
	var dbID, dbHash string
	err := r.DB.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&dbID, &dbHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", errors.New("user not found")
		}
		return "", "", err
	}
	return dbID, dbHash, nil
}
