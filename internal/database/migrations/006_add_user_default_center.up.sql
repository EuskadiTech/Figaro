-- Migration: Add default center fields to users table
-- Version: 006

-- Add default center fields to users table
ALTER TABLE users ADD COLUMN default_center_id INTEGER REFERENCES centers(id) ON DELETE SET NULL;
ALTER TABLE users ADD COLUMN force_default_center BOOLEAN DEFAULT FALSE;

-- Create index for default_center_id for better performance
CREATE INDEX IF NOT EXISTS idx_users_default_center_id ON users(default_center_id);