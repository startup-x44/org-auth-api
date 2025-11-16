-- Remove is_system column from permissions table
ALTER TABLE permissions DROP COLUMN IF EXISTS is_system;
