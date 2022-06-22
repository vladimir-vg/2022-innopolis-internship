-- name: initialize
CREATE TABLE goroutines (
        -- id is the name of the function that was spawned
        id text,
        packageName text,
        filename text,
        line integer
);
CREATE TABLE spawns (
        parentId text,
        childId text,
        filename text,
        line integer
);


-- I need to compute ordering of the goroutines based on ancestry.
-- I can easily express what I want in Datalog:
--
-- given facts as rows:
--
--         goroutine(_Id).
--         goroutine_ancestry(_ParentId, _ChildId).
--
--         rank_option(Id, N) :-
--                 Id = "main-goroutine",
--                 N = 0.
--         rank_option(Id, N2) :-
--                 goroutine_ancestry(ParentId, Id),
--                 rank_option(ParentId, N1),
--                 successor(N1, N2).
--         rank(Id, max<N>) :-
--                 rank_option(Id, N).
--
--
-- I can express same thing in SQL queries.
--
-- Inspired by Jamie Brandon way of writing datalog as recursive sql queries:
-- https://www.scattered-thoughts.net/writing/implicit-ordering-in-relational-languages

CREATE VIEW prepared_goroutines AS
WITH RECURSIVE
        rank_option(id, n) AS (
                SELECT 'main' AS id, 0 AS n
                UNION ALL
                SELECT spawns.childId AS id, rank_option.n+1 AS n
                FROM spawns
                INNER JOIN rank_option ON rank_option.id = spawns.parentId 
        ),
        rank(id, n) AS (
                SELECT rank_option.id AS id, MAX(rank_option.n) AS n
                FROM rank_option
                GROUP BY id
                ORDER BY n
        )
SELECT * FROM rank;

-- name: insert-goroutine
INSERT INTO goroutines (id, packageName, filename, line)
VALUES (?, ?, ?, ?);

-- name: insert-spawn
INSERT INTO spawns (parentId, childId, filename, line)
VALUES (?, ?, ?, ?);
