-- Phase 14: Policy Engine & Notifications Migration

-- 1. Add policies JSONB column to consents table
ALTER TABLE consents ADD COLUMN IF NOT EXISTS policies JSONB DEFAULT '{}'::jsonb;
