-- Migration to create the guides table
CREATE TABLE IF NOT EXISTS guides (
    guide_id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    name VARCHAR(255) NOT NULL,
    area_code VARCHAR(10) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    gender VARCHAR(50),
    skills TEXT,
    languages TEXT,
    photo_url VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(area_code, phone_number)
);

-- Add index for soft deletion lookup
CREATE INDEX IF NOT EXISTS idx_guides_deleted_at ON guides (deleted_at); 