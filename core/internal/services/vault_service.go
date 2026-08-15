package services

import (
	"context"
	"encoding/hex"
	"log"
	"starvault/core/internal/dtos"
	"starvault/core/internal/repository"
)

type VaultService struct {
	VaultRepo  *repository.VaultRepository
	UserRepo   *repository.UserRepository
	EncryptSvc *EncryptionService
}

func (s *VaultService) getEncryptionService(ctx context.Context, userID string) *EncryptionService {
	_, _, userKey, err := s.UserRepo.GetUserByID(ctx, userID)
	if err == nil && userKey != "" {
		if keyBytes, err := hex.DecodeString(userKey); err == nil && len(keyBytes) == 32 {
			if customSvc, err := NewEncryptionService(keyBytes); err == nil {
				return customSvc
			}
		}
	}
	// Fallback to system master key
	return s.EncryptSvc
}

func (s *VaultService) CreateVaultData(ctx context.Context, userID string, req dtos.CreateVaultDataRequest) (*dtos.VaultDataResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	encSvc := s.getEncryptionService(ctx, userID)

	encryptedData, err := encSvc.Encrypt(ctx, []byte(req.Data))
	if err != nil {
		return nil, err // AES errors are caught before hitting DB
	}

	id, createdAt, err := s.VaultRepo.Create(ctx, userID, req.DataType, encryptedData)
	if err != nil {
		return nil, err
	}

	return &dtos.VaultDataResponse{
		ID:        id,
		DataType:  req.DataType,
		CreatedAt: createdAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *VaultService) GetVaultData(ctx context.Context, id string, userID string) (*dtos.VaultDataResponse, error) {
	record, err := s.VaultRepo.GetByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err // Returns "record not found" on IDOR attempt
	}

	// First try user's key if they have one
	encSvc := s.getEncryptionService(ctx, userID)
	plaintext, err := encSvc.Decrypt(ctx, record.EncryptedPayload)
	if err != nil {
		// Fallback to system master key if custom key fails (to support legacy data before they set a key)
		if encSvc != s.EncryptSvc {
			log.Printf("Custom key decryption failed for VaultData %s (user %s). Falling back to master key.", id, userID)
			plaintext, err = s.EncryptSvc.Decrypt(ctx, record.EncryptedPayload)
		}
		
		if err != nil {
			return nil, err // This returns 500 if the data is crypto-shredded (key changed/deleted)
		}
	}

	return &dtos.VaultDataResponse{
		ID:        record.ID,
		DataType:  record.DataType,
		Data:      string(plaintext),
		CreatedAt: record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

