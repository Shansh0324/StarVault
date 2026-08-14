package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"starvault/core/internal/blockchain"
	"starvault/core/internal/repository"
	"time"
)

type AuditService struct {
	AuditRepo        *repository.AuditRepository
	BlockchainClient *blockchain.Client
}

// LogAccessAttempt logs the access attempt synchronously to the durable queue (PostgreSQL).
func (s *AuditService) LogAccessAttempt(ctx context.Context, userID, appID, action, scope string) {
	// Execute synchronously to ensure chronological ordering in Postgres
	err := s.AuditRepo.LogEvent(ctx, userID, appID, action, scope)
	if err != nil {
		log.Printf("ERROR: Failed to durable queue audit log for user %s, app %s: %v", userID, appID, err)
	}
}

// StartAuditWorker starts the background worker that hashes events.
func (s *AuditService) StartAuditWorker(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	log.Println("AuditWorker: Starting background ledger hashing...")

	for {
		select {
		case <-ctx.Done():
			log.Println("AuditWorker: Shutting down.")
			return
		case <-ticker.C:
			s.processPendingAudits(ctx)
		}
	}
}

func (s *AuditService) processPendingAudits(ctx context.Context) {
	// Fetch up to 50 pending events
	logs, err := s.AuditRepo.FetchPending(ctx, 50)
	if err != nil {
		log.Printf("AuditWorker: Error fetching pending audits: %v", err)
		return
	}

	if len(logs) == 0 {
		return
	}

	for _, l := range logs {
		// Get last hash to chain it
		lastHash, err := s.AuditRepo.GetLastHash(ctx, )
		if err != nil {
			log.Printf("AuditWorker: Error getting last hash: %v", err)
			continue
		}

		// Compute new hash: SHA256(lastHash + ID + UserID + AppID + Action + Scope + Timestamp)
		payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", lastHash, l.ID, l.UserID, l.AppID, l.Action, l.Scope, l.Timestamp)
		
		h := sha256.New()
		h.Write([]byte(payload))
		newHash := hex.EncodeToString(h.Sum(nil))

		// Anchor to Blockchain if configured
		if s.BlockchainClient != nil {
			txHash, err := s.BlockchainClient.AnchorHash(ctx, l.ID, newHash)
			if err != nil {
				log.Printf("AuditWorker: Blockchain anchor failed for %s (will retry): %v", l.ID, err)
				continue // Skip local commit to trigger retry on next tick
			}
			log.Printf("AuditWorker: Anchored on-chain. TxHash: %s", txHash)
		}

		// Commit locally
		err = s.AuditRepo.Commit(ctx, l.ID, newHash)
		if err != nil {
			log.Printf("AuditWorker: Error committing hash for %s locally: %v", l.ID, err)
			continue
		}

		log.Printf("AuditWorker: COMMITTED block %s (hash: %s)", l.ID, newHash[:16]+"...")
	}
}
