UPDATE games SET game_type = 'wolf_play' WHERE game_type = 'wolf';

ALTER TABLE games DROP CONSTRAINT games_game_type_check;
ALTER TABLE games ADD CONSTRAINT games_game_type_check
    CHECK (game_type IN ('match_play', 'team_points', 'wolf_play'));