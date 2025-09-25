-- Migration rollback: Drop all tables
DROP INDEX IF EXISTS idx_activities_is_global;
DROP INDEX IF EXISTS idx_activities_start_datetime;
DROP INDEX IF EXISTS idx_activities_center_id;
DROP INDEX IF EXISTS idx_materials_center_id;
DROP INDEX IF EXISTS idx_classrooms_center_id;
DROP INDEX IF EXISTS idx_center_working_hours_center_id;
DROP INDEX IF EXISTS idx_user_permissions_permission;
DROP INDEX IF EXISTS idx_user_permissions_user_id;

DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS materials;
DROP TABLE IF EXISTS classrooms;
DROP TABLE IF EXISTS center_working_hours;
DROP TABLE IF EXISTS centers;
DROP TABLE IF EXISTS user_permissions;
DROP TABLE IF EXISTS users;