ALTER TABLE players
    ADD COLUMN email TEXT,
    ADD CONSTRAINT players_email_unique UNIQUE (email);