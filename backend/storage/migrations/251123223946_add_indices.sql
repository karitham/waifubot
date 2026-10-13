-- migrate:up
CREATE INDEX collection_character_user_idx ON collection (character_id, user_id);

CREATE INDEX character_wishlist_character_user_idx ON character_wishlist (character_id, user_id);

-- migrate:down
DROP INDEX if EXISTS character_wishlist_character_user_idx;

DROP INDEX if EXISTS collection_character_user_idx;
