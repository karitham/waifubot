-- migrate:up
CREATE TABLE character_wishlist (
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    character_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, character_id)
);

CREATE INDEX character_wishlist_user_idx ON character_wishlist(user_id);
CREATE INDEX character_wishlist_character_idx ON character_wishlist(character_id);

-- migrate:down
DROP TABLE IF EXISTS character_wishlist;