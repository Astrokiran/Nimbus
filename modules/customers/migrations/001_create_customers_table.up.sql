-- Migration to create the customers table
CREATE TABLE IF NOT EXISTS customers (
    customer_id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    area_code VARCHAR(10) NOT NULL,
    mobile_number VARCHAR(20) NOT NULL,
    name VARCHAR(255),
    gender VARCHAR(50),
    date_of_birth DATE,
    time_of_birth TIME, -- Using TIME type, ensure compatibility or use VARCHAR
    place_of_birth VARCHAR(255),
    current_address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    pincode VARCHAR(20),
    UNIQUE (area_code, mobile_number) -- Unique constraint on the combination
);

-- Add index for soft deletion lookup
CREATE INDEX IF NOT EXISTS idx_customers_deleted_at ON customers (deleted_at); 