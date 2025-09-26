-- Migration: Add activity sharing and custom links functionality
-- Version: 004

-- Table for sharing activities with specific centers
CREATE TABLE IF NOT EXISTS activity_shares (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    activity_id INTEGER NOT NULL,
    center_id INTEGER NOT NULL,
    shared_by_center_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    FOREIGN KEY (center_id) REFERENCES centers(id) ON DELETE CASCADE,
    FOREIGN KEY (shared_by_center_id) REFERENCES centers(id) ON DELETE CASCADE,
    UNIQUE(activity_id, center_id)
);

-- Table for custom links associated with activities
CREATE TABLE IF NOT EXISTS activity_custom_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    activity_id INTEGER NOT NULL,
    label VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_activity_shares_activity_id ON activity_shares(activity_id);
CREATE INDEX IF NOT EXISTS idx_activity_shares_center_id ON activity_shares(center_id);
CREATE INDEX IF NOT EXISTS idx_activity_custom_links_activity_id ON activity_custom_links(activity_id);