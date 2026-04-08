CREATE TABLE IF NOT EXISTS tasks (
    id                  TEXT PRIMARY KEY NOT NULL,
    active              INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
    title               TEXT NOT NULL,

    rrule               TEXT, -- if null is a single task

    -- Computed Helpers 

    dtstart             DATETIME,
    until               DATETIME,
    count               INTEGER,
    next_occurrence     DATETIME,

    -- Computed Helpers

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS completions (
    id                  TEXT PRIMARY KEY NOT NULL,
    task_id             TEXT NOT NULL,
    occurrence_date     DATETIME NOT NULL,
    completed_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS schema_migrations (
    version     TEXT PRIMARY KEY,
    applied_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_task_next_occurrence ON tasks (next_occurrence, active);
CREATE INDEX IF NOT EXISTS idx_completion_task      ON completions (task_id, occurrence_date);