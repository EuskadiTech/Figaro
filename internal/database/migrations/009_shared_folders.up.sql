-- Migration: Add shared folders table
-- This allows storing both cloud drive links and local file folders

CREATE TABLE shared_folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    center_id INTEGER NULL,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    type TEXT NOT NULL CHECK (type IN ('local', 'cloud')),
    cloud_url TEXT NULL,
    local_path TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (center_id) REFERENCES centers (id) ON DELETE CASCADE
);

-- Index for better performance
CREATE INDEX idx_shared_folders_center_id ON shared_folders (center_id);
CREATE INDEX idx_shared_folders_type ON shared_folders (type);
CREATE INDEX idx_shared_folders_active ON shared_folders (is_active);