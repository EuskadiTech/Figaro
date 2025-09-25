-- Remove URL columns from activities table
ALTER TABLE activities DROP COLUMN meeting_url;
ALTER TABLE activities DROP COLUMN web_url;