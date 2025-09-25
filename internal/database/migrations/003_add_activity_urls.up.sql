-- Add missing columns to activities table
ALTER TABLE activities ADD COLUMN meeting_url VARCHAR(500);
ALTER TABLE activities ADD COLUMN web_url VARCHAR(500);