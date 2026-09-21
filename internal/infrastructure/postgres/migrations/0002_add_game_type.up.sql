ALTER TABLE games
    ADD COLUMN game_type TEXT NOT NULL DEFAULT 'points_play'
    CHECK (game_type IN ('points_play', 'match_play'));