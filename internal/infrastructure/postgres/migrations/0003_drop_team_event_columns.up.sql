ALTER TABLE games DROP COLUMN team_event_id;
ALTER TABLE games DROP COLUMN event_position;
ALTER TABLE games DROP CONSTRAINT games_game_type_check;
ALTER TABLE games ADD CONSTRAINT games_game_type_check
    CHECK (game_type IN ('match_play', 'team_points', 'wolf'));