-- name: create-tables
CREATE TABLE goroutines (
        id text,
        packageName text,
        filename text,
        line integer
);
CREATE TABLE goroutines_ancestry (
        parentId text,
        childId text
);

-- name: insert-goroutine
INSERT INTO goroutines (id, packageName, filename, line)
VALUES (?, ?, ?, ?);

-- name: insert-goroutine-ancestry
INSERT INTO goroutines_ancestry (parentId, childId)
VALUES (?, ?);
