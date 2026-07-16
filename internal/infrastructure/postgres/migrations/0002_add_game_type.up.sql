ALTER TABLE games
    ADD COLUMN game_type TEXT NOT NULL DEFAULT 'team_play'
    CHECK (game_type IN ('team_play', 'match_play'));