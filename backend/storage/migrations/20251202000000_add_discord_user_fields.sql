-- migrate:up
-- Add Discord user information to users table
ALTER TABLE users 
ADD COLUMN "discord_username" VARCHAR(32) NOT NULL DEFAULT '',
ADD COLUMN "discord_avatar" VARCHAR(34) NOT NULL DEFAULT '',
ADD COLUMN "last_updated" TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:00-00';

-- migrate:down
-- Remove Discord user information from users table
ALTER TABLE users 
DROP COLUMN IF EXISTS "discord_username",
DROP COLUMN IF EXISTS "discord_avatar",
DROP COLUMN IF EXISTS "last_updated";