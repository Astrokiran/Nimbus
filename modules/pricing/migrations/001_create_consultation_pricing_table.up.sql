-- Migration UP: Create the consultation_pricing table
CREATE TABLE IF NOT EXISTS consultation_pricing (
    id SERIAL PRIMARY KEY,                       -- Changed BIGSERIAL to SERIAL to match guides table
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,                   -- Keeping BIGINT assuming guide/customer IDs might exceed INT range, adjust if known otherwise
    chat_rate_per_min BIGINT DEFAULT 0,
    call_rate_per_min BIGINT DEFAULT 0,
    video_call_rate_per_min BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    UNIQUE(entity_type, entity_id)             -- Simplified UNIQUE constraint syntax
);

-- Add index for soft deletion lookup
CREATE INDEX IF NOT EXISTS idx_consultation_pricing_deleted_at ON consultation_pricing (deleted_at);

-- Optional: Add index on entity_type and entity_id for faster lookups if needed beyond UNIQUE constraint
-- CREATE INDEX IF NOT EXISTS idx_consultation_pricing_entity ON consultation_pricing (entity_type, entity_id); 