package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"starvault/core/internal/dtos"
	"starvault/core/internal/repository"
)

type AppService struct {
	AppRepo *repository.AppRepository
}

func (s *AppService) CreateApp(ctx context.Context, req dtos.CreateAppRequest) (*dtos.CreateAppResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Generate a 32-byte cryptographically secure random secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, errors.New("failed to generate secure application secret")
	}

	// Base64Url encode it so it's clean for HTTP headers
	plaintextSecret := base64.RawURLEncoding.EncodeToString(secretBytes)

	// Hash the secret for storage
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash application secret")
	}

	// Store only the hash
	appID, err := s.AppRepo.CreateApp(ctx, req.Name, string(hash))
	if err != nil {
		return nil, err
	}

	// Return the ID and plaintext secret ONCE
	return &dtos.CreateAppResponse{
		AppID:  appID,
		Name:   req.Name,
		Secret: plaintextSecret,
	}, nil
}
