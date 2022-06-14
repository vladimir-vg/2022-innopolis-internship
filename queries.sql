-- name: create-tables
CREATE TABLE goroutines (
        id text,
        package_name text,
        filename text,
        line integer,
        parent_id text
);
CREATE TABLE goroutines_ancestry (
        parent_id text,
        child_id text
);

-- name: insert-goroutine
INSERT INTO goroutines (id, package_name, func_Name, filename, line)
VALUES (?, ?, ?, ?);

-- name: insert-goroutine-ancestry
INSERT INTO goroutines_ancestry (parent_id, child_id)
VALUES (?, ?);
