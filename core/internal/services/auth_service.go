package services

import (
	"errors"
	"strings"
	"starvault/core/internal/dtos"
	"starvault/core/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo *repository.UserRepository
}

func (s *AuthService) Register(req dtos.AuthRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("Failed to hash password")
	}

	id, err := s.UserRepo.CreateUser(req.Email, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return "", errors.New("Email already exists")
		}
		return "", err
	}
	return id, nil
}

func (s *AuthService) Login(req dtos.AuthRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", err
	}

	dbID, dbHash, err := s.UserRepo.GetUserByEmail(req.Email)
	if err != nil {
		return "", errors.New("Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbHash), []byte(req.Password)); err != nil {
		return "", errors.New("Invalid email or password")
	}

	return dbID, nil
}
