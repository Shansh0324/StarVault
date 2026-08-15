package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"starvault/core/internal/dtos"
	"starvault/core/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo    *repository.UserRepository
	RiskService *RiskService
}

func (s *AuthService) Register(ctx context.Context, req dtos.AuthRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("Failed to hash password")
	}

	id, err := s.UserRepo.CreateUser(ctx, req.Email, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return "", errors.New("Email already exists")
		}
		return "", err
	}

	// Generate and set DID
	did := fmt.Sprintf("did:starvault:%s", id)
	err = s.UserRepo.SetDID(ctx, id, did)
	if err != nil {
		log.Printf("Failed to set DID for user %s: %v", id, err)
	}

	return id, nil
}

func (s *AuthService) Login(ctx context.Context, req dtos.AuthRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", err
	}

	dbID, dbHash, err := s.UserRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return "", errors.New("Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbHash), []byte(req.Password)); err != nil {
		return "", errors.New("Invalid email or password")
	}

	// Evaluate Device Posture / Risk Score
	if s.RiskService != nil {
		posture := DevicePosture{
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
		}
		riskScore, err := s.RiskService.EvaluateLogin(ctx, dbID, posture)
		if err != nil {
			log.Printf("RiskService error for user %s: %v", dbID, err)
		} else if riskScore >= RiskLevelHigh {
			log.Printf("HIGH RISK LOGIN for user %s from IP %s", dbID, req.IPAddress)
			// For now, we only log it. We don't block access yet to avoid locking users out.
		}
	}

	return dbID, nil
}
