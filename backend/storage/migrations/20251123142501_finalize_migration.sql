-- migrate:up
-- Rename old table to backup
ALTER TABLE characters RENAME TO characters_backup;

-- Rename new tables to final names
ALTER TABLE characters_new RENAME TO characters;
-- collection already has correct name

-- Update users table favorite constraint to work with new structure
-- Since favorite references a character ID but we need to know which user owns it
-- For now, drop the constraint and handle validation in application logic
ALTER TABLE users
DROP CONSTRAINT IF EXISTS "characters_users_fk";

-- Create index on characters table for search
CREATE INDEX characters_name_idx ON characters USING gin(to_tsvector('english', name));

-- migrate:down
-- Restore original structure
DROP TABLE IF EXISTS characters;
ALTER TABLE characters_backup RENAME TO characters;

-- Drop new table
DROP TABLE IF EXISTS collection;

-- Restore original foreign key
ALTER TABLE users
ADD CONSTRAINT "characters_users_fk" 
FOREIGN KEY (favorite, user_id) 
REFERENCES characters(id, user_id);