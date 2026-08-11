package services

import (
	"context"
	"starvault/core/internal/dtos"
	"starvault/core/internal/repository"
)

type VaultService struct {
	VaultRepo  *repository.VaultRepository
	EncryptSvc *EncryptionService
}

func (s *VaultService) CreateVaultData(ctx context.Context, userID string, req dtos.CreateVaultDataRequest) (*dtos.VaultDataResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	encryptedData, err := s.EncryptSvc.Encrypt(ctx, []byte(req.Data))
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

	plaintext, err := s.EncryptSvc.Decrypt(ctx, record.EncryptedPayload)
	if err != nil {
		return nil, err
	}

	return &dtos.VaultDataResponse{
		ID:        record.ID,
		DataType:  record.DataType,
		Data:      string(plaintext),
		CreatedAt: record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
