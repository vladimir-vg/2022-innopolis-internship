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
-- I can easily express what I want in Datalog and then
-- write corresponding recursive SQL query.
--
--         ancestry_rank_option(Id, Rank) :-
--                 Id = "main-goroutine",
--                 Rank = 0.
--         ancestry_rank_option(Id, Rank2) :-
--                 spawns(ParentId, Id),
--                 ancestry_rank_option(ParentId, Rank1),
--                 Rank2 = Rank1 + 1.
--         ancestry_rank(Id, max<Rank>) :-
--                 ancestry_rank_option(Id, Rank).
--
-- Inspired by Jamie Brandon way of writing datalog as recursive sql queries:
-- https://www.scattered-thoughts.net/writing/implicit-ordering-in-relational-languages

CREATE VIEW ancestry_rank AS
WITH RECURSIVE
        ancestry_rank_option(id, rank) AS (
                SELECT 'main' AS id, 0 AS rank
                UNION ALL
                SELECT spawns.childId AS id, ancestry_rank_option.rank+1 AS rank
                FROM spawns
                INNER JOIN ancestry_rank_option ON ancestry_rank_option.id = spawns.parentId 
        )
SELECT ancestry_rank_option.id AS id, MAX(ancestry_rank_option.rank) AS rank
FROM ancestry_rank_option
GROUP BY id
ORDER BY rank;

-- CREATE VIEW time_events AS
-- SELECT
--         'goroutine-start' AS type,
--         filename,
--         line,
--         id
-- FROM goroutines
-- INNER JOIN ancestry_rank ON ancestry_rank.id = goroutines.id

-- name: insert-goroutine
INSERT INTO goroutines (id, packageName, filename, line)
VALUES (?, ?, ?, ?);

-- name: insert-spawn
INSERT INTO spawns (parentId, childId, filename, line)
VALUES (?, ?, ?, ?);
