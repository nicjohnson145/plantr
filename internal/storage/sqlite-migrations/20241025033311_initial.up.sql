BEGIN;

CREATE TABLE challenge (
    id    TEXT NOT NULL PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE github_release_version (
    repo           TEXT NOT NULL PRIMARY KEY,
    latest_version TEXT NOT NULL,
    last_check     DATETIME
);

COMMIT;
