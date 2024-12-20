BEGIN;

CREATE TABLE github_release_cache (
    repo         TEXT NOT NULL,
    tag          TEXT NOT NULL,
    os           TEXT NOT NULL,
    arch         TEXT NOT NULL,
    download_url TEXT NOT NULL,
    PRIMARY KEY (repo, tag, os, arch)
);

COMMIT;
