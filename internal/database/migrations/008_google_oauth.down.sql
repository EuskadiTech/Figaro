-- Remove Google OAuth settings from system_settings table
DELETE FROM system_settings WHERE category = 'oauth';