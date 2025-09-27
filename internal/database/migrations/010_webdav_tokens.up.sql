-- Migration: Add WebDAV device session tokens
-- This allows device-based WebDAV authentication with named sessions

CREATE TABLE webdav_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    device_name TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP DEFAULT (datetime('now', '+30 days')),
    is_active BOOLEAN NOT NULL DEFAULT 1,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Index for performance
CREATE INDEX idx_webdav_tokens_token ON webdav_tokens (token);
CREATE INDEX idx_webdav_tokens_user_id ON webdav_tokens (user_id);
CREATE INDEX idx_webdav_tokens_active ON webdav_tokens (is_active, expires_at);