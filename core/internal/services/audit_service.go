package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"starvault/core/internal/blockchain"
	"starvault/core/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type AuditService struct {
	AuditRepo        *repository.AuditRepository
	BlockchainClient *blockchain.Client
	JetStream        nats.JetStreamContext
	BatchInterval    time.Duration // default 60s
	BatchMaxSize     int           // default 1000
}

type AuditEvent struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	AppID     string `json:"appId"`
	Action    string `json:"action"`
	Scope     string `json:"scope"`
	Timestamp string `json:"timestamp"`
}

// LogAccessAttempt publishes the access attempt to NATS JetStream.
func (s *AuditService) LogAccessAttempt(ctx context.Context, userID, appID, action, scope string) {
	event := AuditEvent{
		ID:        uuid.New().String(),
		UserID:    userID,
		AppID:     appID,
		Action:    action,
		Scope:     scope,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("ERROR: Failed to serialize audit event: %v", err)
		return
	}

	if s.JetStream != nil {
		_, err = s.JetStream.Publish("audit.events", payload)
		if err != nil {
			log.Printf("ERROR: Failed to publish audit event to NATS: %v", err)
		}
	} else {
		s.AuditRepo.LogEvent(ctx, userID, appID, action, scope)
	}
}

// StartAuditWorker subscribes to NATS JetStream and inserts events to DB.
// This is the FAST PATH — no blockchain calls, just hash + DB insert + Ack.
func (s *AuditService) StartAuditWorker(ctx context.Context) {
	if s.JetStream == nil {
		log.Println("AuditWorker: JetStream not configured. Exiting worker.")
		return
	}

	log.Println("AuditWorker: Subscribing to NATS JetStream for audit.events...")

	sub, err := s.JetStream.Subscribe("audit.events", func(m *nats.Msg) {
		var event AuditEvent
		if err := json.Unmarshal(m.Data, &event); err != nil {
			log.Printf("AuditWorker: Error decoding NATS message: %v", err)
			m.Nak()
			return
		}

		s.processEvent(ctx, event, m)
	}, nats.Durable("AUDIT_WORKER"), nats.ManualAck())

	if err != nil {
		log.Fatalf("AuditWorker: Failed to subscribe to JetStream: %v", err)
	}
	defer sub.Unsubscribe()

	<-ctx.Done()
	log.Println("AuditWorker: Shutting down.")
}

// processEvent hashes and inserts the event to DB as PENDING_BATCH, then Acks.
func (s *AuditService) processEvent(ctx context.Context, event AuditEvent, m *nats.Msg) {
	lastHash, err := s.AuditRepo.GetLastHash(ctx)
	if err != nil {
		log.Printf("AuditWorker: Error getting last hash: %v", err)
		m.Nak()
		return
	}

	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", lastHash, event.ID, event.UserID, event.AppID, event.Action, event.Scope, event.Timestamp)
	h := sha256.New()
	h.Write([]byte(payload))
	newHash := hex.EncodeToString(h.Sum(nil))

	err = s.AuditRepo.InsertPendingBatch(ctx, event.ID, event.UserID, event.AppID, event.Action, event.Scope, newHash)
	if err != nil {
		log.Printf("AuditWorker: Error inserting event %s: %v", event.ID, err)
		m.Nak()
		return
	}

	m.Ack()
	log.Printf("AuditWorker: Inserted event %s (hash: %s...)", event.ID, newHash[:16])
}

// StartBatchWorker runs on a ticker and batches PENDING_BATCH events into Merkle trees.
// One blockchain transaction per batch instead of one per event.
func (s *AuditService) StartBatchWorker(ctx context.Context) {
	interval := s.BatchInterval
	if interval == 0 {
		interval = 60 * time.Second
	}
	maxSize := s.BatchMaxSize
	if maxSize == 0 {
		maxSize = 1000
	}

	log.Printf("BatchWorker: Started (interval=%s, maxSize=%d)", interval, maxSize)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("BatchWorker: Shutting down.")
			return
		case <-ticker.C:
			s.processBatch(ctx, maxSize)
		}
	}
}

// processBatch fetches pending events, builds a Merkle tree, anchors root on-chain, commits.
func (s *AuditService) processBatch(ctx context.Context, maxSize int) {
	events, err := s.AuditRepo.FetchPendingBatch(ctx, maxSize)
	if err != nil {
		log.Printf("BatchWorker: Error fetching pending events: %v", err)
		return
	}
	if len(events) == 0 {
		return // nothing to batch
	}

	// Build Merkle tree from event hashes
	leaves := make([][]byte, len(events))
	eventIDs := make([]string, len(events))
	for i, e := range events {
		hashBytes, _ := hex.DecodeString(e.EventHash)
		leaves[i] = hashBytes
		eventIDs[i] = e.ID
	}

	tree := &blockchain.MerkleTree{Leaves: leaves}
	merkleRoot := hex.EncodeToString(tree.Root())
	batchID := uuid.New().String()

	log.Printf("BatchWorker: Built Merkle tree for %d events (root: %s...)", len(events), merkleRoot[:16])

	// Anchor on blockchain (single transaction)
	var txHash string
	if s.BlockchainClient != nil {
		bcCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		tx, err := s.BlockchainClient.AnchorHash(bcCtx, batchID, merkleRoot)
		if err != nil {
			log.Printf("BatchWorker: Blockchain anchor failed for batch %s: %v (will retry next tick)", batchID, err)
			return // events stay PENDING_BATCH, retry next tick
		}
		txHash = tx
		log.Printf("BatchWorker: Anchored batch on-chain. TxHash: %s", txHash)
	} else {
		txHash = "no-blockchain"
	}

	// Commit batch to DB (atomic: creates batch record + updates all events)
	committedBatchID, err := s.AuditRepo.CreateBatchAndCommit(ctx, merkleRoot, txHash, eventIDs)
	if err != nil {
		log.Printf("BatchWorker: Error committing batch to DB: %v", err)
		return
	}

	log.Printf("BatchWorker: COMMITTED batch %s (%d events, root: %s...)", committedBatchID, len(events), merkleRoot[:16])
}
