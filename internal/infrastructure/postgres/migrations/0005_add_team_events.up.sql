CREATE TABLE team_events (
    id          TEXT PRIMARY KEY,
    course_id   TEXT NOT NULL,
    course_name TEXT NOT NULL,
    variant     TEXT NOT NULL CHECK (variant IN ('gross', 'net')),
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMP NULL
);

ALTER TABLE games
    ADD COLUMN team_event_id  TEXT NULL REFERENCES team_events(id) ON DELETE SET NULL,
    ADD COLUMN event_position INTEGER NULL;