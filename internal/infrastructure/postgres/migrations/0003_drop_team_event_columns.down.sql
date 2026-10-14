ALTER TABLE games DROP CONSTRAINT games_game_type_check;
ALTER TABLE games ADD CONSTRAINT games_game_type_check
    CHECK (game_type IN ('points_play', 'match_play', 'wolf', 'team_event'));

UPDATE games SET game_type = 'wolf' WHERE game_type = 'wolf_play';