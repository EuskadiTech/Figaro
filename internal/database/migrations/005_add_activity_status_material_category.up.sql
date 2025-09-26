-- Migration: Add activity status and material category fields
-- Version: 005

-- Add status column to activities table
ALTER TABLE activities ADD COLUMN status VARCHAR(20) DEFAULT 'pending';

-- Add category column to materials table  
ALTER TABLE materials ADD COLUMN category VARCHAR(100) DEFAULT 'general';

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_activities_status ON activities(status);
CREATE INDEX IF NOT EXISTS idx_materials_category ON materials(category);

-- Update existing records to have proper status based on dates
UPDATE activities SET status = 'completed' WHERE end_datetime < datetime('now');
UPDATE activities SET status = 'in_progress' WHERE start_datetime <= datetime('now') AND end_datetime >= datetime('now');
UPDATE activities SET status = 'pending' WHERE start_datetime > datetime('now');