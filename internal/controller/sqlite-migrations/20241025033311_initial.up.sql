BEGIN;

CREATE TABLE challenge (
    id    TEXT NOT NULL PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE github_release_asset (
    hash         TEXT NOT NULL,
    os           TEXT NOT NULL,
    arch         TEXT NOT NULL,
    download_url TEXT NOT NULL,
    PRIMARY KEY (hash, os, arch)
);

COMMIT;
