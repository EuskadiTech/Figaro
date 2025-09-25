-- Migration: Create users table
-- Version: 001

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    permission VARCHAR(255) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, permission)
);

CREATE TABLE IF NOT EXISTS centers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) UNIQUE NOT NULL,
    timezone VARCHAR(50) DEFAULT 'Europe/Madrid',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS center_working_hours (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    center_id INTEGER NOT NULL,
    day_of_week INTEGER NOT NULL, -- 0=Sunday, 1=Monday, ..., 6=Saturday
    start_time TIME,
    end_time TIME,
    FOREIGN KEY (center_id) REFERENCES centers(id) ON DELETE CASCADE,
    UNIQUE(center_id, day_of_week)
);

CREATE TABLE IF NOT EXISTS classrooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    center_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (center_id) REFERENCES centers(id) ON DELETE CASCADE,
    UNIQUE(center_id, name)
);

CREATE TABLE IF NOT EXISTS materials (
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

CREATE TABLE IF NOT EXISTS activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    center_id INTEGER NULL, -- NULL for global activities
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_datetime DATETIME NOT NULL,
    end_datetime DATETIME NOT NULL,
    is_global BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (center_id) REFERENCES centers(id) ON DELETE CASCADE
);

-- Insert default demo user
INSERT OR IGNORE INTO users (username, password_hash, display_name, email) 
VALUES ('demo', '$2y$10$wonKfVlZndi4K22lDV5Ce.QtWGIvmnOPxrJ1jOmWPUCUgGBw1Waku', 'Demo User', 'demo@example.com');

-- Insert default permissions for demo user
INSERT OR IGNORE INTO user_permissions (user_id, permission) 
SELECT id, 'ADMIN' FROM users WHERE username = 'demo';

INSERT OR IGNORE INTO user_permissions (user_id, permission) 
SELECT id, 'module2' FROM users WHERE username = 'demo';

-- Insert default centers
INSERT OR IGNORE INTO centers (name) VALUES ('Centro Demo');
INSERT OR IGNORE INTO centers (name) VALUES ('Centro Demo 2');

-- Insert working hours for Centro Demo
INSERT OR IGNORE INTO center_working_hours (center_id, day_of_week, start_time, end_time)
SELECT id, 1, '08:00', '18:00' FROM centers WHERE name = 'Centro Demo';
INSERT OR IGNORE INTO center_working_hours (center_id, day_of_week, start_time, end_time)
SELECT id, 2, '08:00', '18:00' FROM centers WHERE name = 'Centro Demo';
INSERT OR IGNORE INTO center_working_hours (center_id, day_of_week, start_time, end_time)
SELECT id, 3, '08:00', '18:00' FROM centers WHERE name = 'Centro Demo';
INSERT OR IGNORE INTO center_working_hours (center_id, day_of_week, start_time, end_time)
SELECT id, 4, '08:00', '18:00' FROM centers WHERE name = 'Centro Demo';
INSERT OR IGNORE INTO center_working_hours (center_id, day_of_week, start_time, end_time)
SELECT id, 5, '08:00', '18:00' FROM centers WHERE name = 'Centro Demo';

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_permissions_permission ON user_permissions(permission);
CREATE INDEX IF NOT EXISTS idx_center_working_hours_center_id ON center_working_hours(center_id);
CREATE INDEX IF NOT EXISTS idx_classrooms_center_id ON classrooms(center_id);
CREATE INDEX IF NOT EXISTS idx_materials_center_id ON materials(center_id);
CREATE INDEX IF NOT EXISTS idx_activities_center_id ON activities(center_id);
CREATE INDEX IF NOT EXISTS idx_activities_start_datetime ON activities(start_datetime);
CREATE INDEX IF NOT EXISTS idx_activities_is_global ON activities(is_global);