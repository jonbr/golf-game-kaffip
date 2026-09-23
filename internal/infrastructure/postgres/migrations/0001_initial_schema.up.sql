-- ============================
-- Players
-- ============================
CREATE TABLE players (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    email       TEXT NOT NULL,
    handicap    DOUBLE PRECISION NOT NULL,
    deleted_at  TIMESTAMP NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT players_email_unique UNIQUE (email)
);

CREATE INDEX idx_players_deleted_at
    ON players (deleted_at);

-- ============================
-- Games
-- Shared header for every game type: points_play, match_play, wolf,
-- and team_event (which aggregates other games via team_event_id).
-- variant/starting_lead/match_team_a/match_team_b are only meaningful
-- for points_play/match_play; wolf and team_event rows leave them at
-- their defaults.
-- ============================
CREATE TABLE games (
    id             TEXT PRIMARY KEY,
    game_type      TEXT NOT NULL DEFAULT 'points_play'
                   CHECK (game_type IN ('points_play', 'match_play', 'wolf', 'team_event')),
    course_id      TEXT NOT NULL,
    course_name    TEXT NOT NULL,
    variant        TEXT NOT NULL DEFAULT 'gross' CHECK (variant IN ('gross', 'net')),
    starting_lead  INTEGER NOT NULL DEFAULT 0,
    current_hole   INTEGER NOT NULL DEFAULT 1,
    match_team_a   INTEGER NOT NULL DEFAULT 0,
    match_team_b   INTEGER NOT NULL DEFAULT 0,
    team_event_id  TEXT NULL REFERENCES games(id) ON DELETE SET NULL,
    event_position INTEGER NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at    TIMESTAMP NULL
);

-- ============================
-- Game Players (join table)
-- team is used by points_play/match_play/team_event matches ('A'/'B').
-- seat is used by wolf's fixed rotation order (0-3). Exactly one of the
-- two is populated depending on game_type.
-- ============================
CREATE TABLE game_players (
    game_id     TEXT NOT NULL,
    player_id   BIGINT NOT NULL,
    team        CHAR(1) NULL CHECK (team IN ('A', 'B')),
    seat        INTEGER NULL CHECK (seat BETWEEN 0 AND 3),
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

-- ============================
-- Wolf Hole Results
-- One row per wolf game + hole, upserted on correction. Not part of the
-- shared hole_results table since Wolf's shape (individual players,
-- rotating mode/partner) doesn't fit the two-sided points_a/points_b
-- columns used by points_play/match_play.
-- ============================
CREATE TABLE wolf_hole_results (
    id             BIGSERIAL PRIMARY KEY,
    game_id        TEXT NOT NULL,
    hole_number    INTEGER NOT NULL,
    wolf_player_id BIGINT NOT NULL,
    mode           TEXT NOT NULL CHECK (mode IN ('partnered', 'lone')),
    partner_id     BIGINT NULL,
    winning_side   TEXT NULL CHECK (winning_side IN ('wolf', 'field')),
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE (game_id, hole_number),

    FOREIGN KEY (game_id)
        REFERENCES games(id)
        ON DELETE CASCADE,

    FOREIGN KEY (wolf_player_id)
        REFERENCES players(id)
        ON DELETE RESTRICT,

    FOREIGN KEY (partner_id)
        REFERENCES players(id)
        ON DELETE RESTRICT
);

-- ============================
-- Wolf Hole Scores
-- All 4 players' scores + points for a given wolf hole result.
-- ============================
CREATE TABLE wolf_hole_scores (
    id                  BIGSERIAL PRIMARY KEY,
    wolf_hole_result_id BIGINT NOT NULL,
    player_id           BIGINT NOT NULL,
    gross               INTEGER NOT NULL,
    net                 INTEGER NOT NULL,
    strokes             INTEGER NOT NULL,
    points              INTEGER NOT NULL DEFAULT 0,

    UNIQUE (wolf_hole_result_id, player_id),

    FOREIGN KEY (wolf_hole_result_id)
        REFERENCES wolf_hole_results(id)
        ON DELETE CASCADE,

    FOREIGN KEY (player_id)
        REFERENCES players(id)
        ON DELETE RESTRICT
);