-- Migration: Add Merkle batch anchoring support
-- This migration adds a batching table for Merkle root anchoring
-- and modifies audit_logs to reference batches.

-- Table to track Merkle batches anchored on-chain
CREATE TABLE IF NOT EXISTS audit_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merkle_root TEXT NOT NULL,
    tx_hash TEXT,
    event_count INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ANCHORED', 'FAILED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    anchored_at TIMESTAMP WITH TIME ZONE
);

-- Add batch reference and index to audit_logs
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS batch_id UUID REFERENCES audit_batches(id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_blockchain_status ON audit_logs(blockchain_status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_batch_id ON audit_logs(batch_id);
