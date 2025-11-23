-- migrate:up
-- Create new normalized characters table
CREATE TABLE characters_new (
    id BIGINT PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    image VARCHAR(256) NOT NULL
);

-- Create new collection table
CREATE TABLE collection (
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    character_id BIGINT NOT NULL REFERENCES characters_new(id),
    source VARCHAR(50) NOT NULL DEFAULT 'ROLL',
    acquired_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, character_id)
);

-- Populate new characters table with unique characters from existing data
-- Use the name and image from the first occurrence of each character ID
INSERT INTO characters_new (id, name, image)
SELECT DISTINCT ON (id)
    id,
    name,
    image
FROM characters
ORDER BY id, date; -- Use the earliest entry for each character

-- Populate new collection table with existing ownership data
INSERT INTO collection (user_id, character_id, source, acquired_at)
SELECT 
    user_id,
    id as character_id,
    COALESCE(type, 'ROLL') as source,
    date as acquired_at
FROM characters;

-- Create indexes for new tables
CREATE INDEX collection_user_idx ON collection(user_id, acquired_at DESC);
CREATE INDEX collection_character_idx ON collection(character_id);

-- migrate:down
-- Drop new tables
DROP TABLE IF EXISTS collection;
DROP TABLE IF EXISTS characters_new;