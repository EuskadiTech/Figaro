-- Revert migration: Add activity status and material category fields
-- Version: 005

-- Remove indexes
DROP INDEX IF EXISTS idx_activities_status;
DROP INDEX IF EXISTS idx_materials_category;

-- Remove columns (SQLite doesn't support DROP COLUMN directly, so we need to recreate tables)
-- For activities table
CREATE TABLE activities_temp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    center_id INTEGER NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_datetime DATETIME NOT NULL,
    end_datetime DATETIME NOT NULL,
    is_global BOOLEAN DEFAULT FALSE,
    meeting_url VARCHAR(500),
    web_url VARCHAR(500),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (center_id) REFERENCES centers(id) ON DELETE CASCADE
);

INSERT INTO activities_temp SELECT id, center_id, title, description, start_datetime, end_datetime, is_global, meeting_url, web_url, created_at, updated_at FROM activities;
DROP TABLE activities;
ALTER TABLE activities_temp RENAME TO activities;

-- For materials table
CREATE TABLE materials_temp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    center_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    photo_path VARCHAR(500),
    unit VARCHAR(100) NOT NULL,
    available_quantity INTEGER NOT NULL DEFAULT 0,
    minimum_quantity INTEGER NOT NULL DEFAULT 0,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (center_id) REFERENCES centers(id) ON DELETE CASCADE
);

INSERT INTO materials_temp SELECT id, center_id, name, photo_path, unit, available_quantity, minimum_quantity, notes, created_at, updated_at FROM materials;
DROP TABLE materials;
ALTER TABLE materials_temp RENAME TO materials;

-- Recreate original indexes
CREATE INDEX IF NOT EXISTS idx_materials_center_id ON materials(center_id);
CREATE INDEX IF NOT EXISTS idx_activities_center_id ON activities(center_id);
CREATE INDEX IF NOT EXISTS idx_activities_start_datetime ON activities(start_datetime);
CREATE INDEX IF NOT EXISTS idx_activities_is_global ON activities(is_global);