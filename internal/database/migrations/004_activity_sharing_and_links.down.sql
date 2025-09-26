-- Rollback: Remove activity sharing and custom links functionality
-- Version: 004

DROP TABLE IF EXISTS activity_custom_links;
DROP TABLE IF EXISTS activity_shares;