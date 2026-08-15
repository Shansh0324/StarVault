-- Partial index for the hot consent lookup query (CheckConsent).
-- Covers only ACTIVE rows, which is exactly what GetActiveConsentForApp filters on.
CREATE INDEX IF NOT EXISTS idx_consents_active_lookup
ON consents (user_id, app_id) WHERE status = 'ACTIVE';
