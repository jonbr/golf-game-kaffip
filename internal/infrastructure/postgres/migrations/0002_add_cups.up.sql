CREATE TABLE cups (
    id          TEXT PRIMARY KEY,
    name        TEXT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMP NULL
);

CREATE TABLE cup_players (
    cup_id    TEXT NOT NULL REFERENCES cups(id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players(id),
    side      CHAR(1) NOT NULL CHECK (side IN ('A', 'B')),
    PRIMARY KEY (cup_id, player_id)
);

CREATE TABLE cup_matches (
    cup_id       TEXT NOT NULL REFERENCES cups(id) ON DELETE CASCADE,
    game_id      TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    cup_position INTEGER NOT NULL,
    PRIMARY KEY (cup_id, game_id)
);