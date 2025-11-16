-- migrate:up
ALTER TABLE characters
DROP CONSTRAINT users_characters_fk;

-- migrate:down
ALTER TABLE characters
ADD CONSTRAINT users_characters_fk FOREIGN KEY (user_id) REFERENCES users (user_id) ON UPDATE NO ACTION ON DELETE NO ACTION;
