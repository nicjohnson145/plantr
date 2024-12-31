BEGIN;

CREATE TABLE inventory_config_file (
    path TEXT NOT NULL,
    PRIMARY KEY (path)
);

COMMIT;
