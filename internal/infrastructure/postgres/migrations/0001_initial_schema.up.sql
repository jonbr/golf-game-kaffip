-- ============================
-- Players
-- ============================
CREATE TABLE players (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    handicap    DOUBLE PRECISION NOT NULL,
    deleted_at  TIMESTAMP NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_players_deleted_at
    ON players (deleted_at);

-- ============================
-- Games
-- ============================
CREATE TABLE games (
    id            TEXT PRIMARY KEY,
    course_id     TEXT NOT NULL,
    course_name   TEXT NOT NULL,
    variant       TEXT NOT NULL DEFAULT 'gross' CHECK (variant IN ('gross', 'net')),
    starting_lead INTEGER NOT NULL DEFAULT 0,
    current_hole  INTEGER NOT NULL DEFAULT 1,
    match_team_a  INTEGER NOT NULL DEFAULT 0,
    match_team_b  INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at   TIMESTAMP NULL
);

-- ============================
-- Game Players (join table)
-- ============================
CREATE TABLE game_players (
    game_id     TEXT NOT NULL,
    player_id   BIGINT NOT NULL,
    team        CHAR(1) NOT NULL CHECK (team IN ('A', 'B')),
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (game_id, player_id),

    FOREIGN KEY (game_id)
        REFERENCES games(id)
        ON DELETE RESTRICT,

    FOREIGN KEY (player_id)
        REFERENCES players(id)
        ON DELETE RESTRICT
);

-- ============================
-- Course hole snapshot
-- Populated once at CreateGame time from the external API response,
-- so GetGame/GetGames/SetHoleScore never need to call it again.
-- ============================
CREATE TABLE game_course_holes (
    game_id        TEXT NOT NULL,
    hole_number    INTEGER NOT NULL,
    par            INTEGER NOT NULL,
    handicap_index INTEGER NOT NULL,

    PRIMARY KEY (game_id, hole_number),

    FOREIGN KEY (game_id)
        REFERENCES games(id)
        ON DELETE CASCADE
);

-- ============================
-- Hole Results
-- One row per game+hole, upserted on correction.
-- ============================
CREATE TABLE hole_results (
    id                     BIGSERIAL PRIMARY KEY,
    game_id                TEXT NOT NULL,
    hole_number            INTEGER NOT NULL,
    points_a               INTEGER NOT NULL,
    points_b               INTEGER NOT NULL,
    low_score_winner_team  CHAR(1) NULL CHECK (low_score_winner_team IN ('A', 'B')),
    team_total_winner_team CHAR(1) NULL CHECK (team_total_winner_team IN ('A', 'B')),
    created_at             TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE (game_id, hole_number),

    FOREIGN KEY (game_id)
        REFERENCES games(id)
        ON DELETE CASCADE
);

-- ============================
-- Per-player hole scores
-- ============================
CREATE TABLE hole_result_scores (
    id             BIGSERIAL PRIMARY KEY,
    hole_result_id BIGINT NOT NULL,
    player_id      BIGINT NOT NULL,
    gross          INTEGER NOT NULL,
    net            INTEGER NOT NULL,
    strokes        INTEGER NOT NULL,
    gross_bonus    INTEGER NOT NULL DEFAULT 0,

    UNIQUE (hole_result_id, player_id),

    FOREIGN KEY (hole_result_id)
        REFERENCES hole_results(id)
        ON DELETE CASCADE,

    FOREIGN KEY (player_id)
        REFERENCES players(id)
        ON DELETE RESTRICT
);