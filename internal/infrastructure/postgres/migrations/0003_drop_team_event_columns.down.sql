ALTER TABLE games DROP CONSTRAINT games_game_type_check;
ALTER TABLE games ADD CONSTRAINT games_game_type_check
    CHECK (game_type IN ('points_play', 'match_play', 'wolf', 'team_event'));

ALTER TABLE games ADD COLUMN team_event_id TEXT NULL REFERENCES games(id) ON DELETE SET NULL;
ALTER TABLE games ADD COLUMN event_position INTEGER NULL;