BEGIN;

CREATE TABLE agent_inventory ( 
    hash     TEXT NOT NULL,
    path     TEXT,
    package  TEXT
    PRIMARY KEY (hash)
);

CREATE UNIQUE INDEX
    only_one_path 
ON
    agent_inventory
    (
        path
    )
;

CREATE UNIQUE INDEX
    only_one_package
ON
    agent_inventory
    (
        package
    )
;

COMMIT;
