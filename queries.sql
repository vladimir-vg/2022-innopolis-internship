-- name: initialize
CREATE TABLE goroutines (
        -- id is the name of the function that was spawned
        id text NOT NULL,
        packageName text NOT NULL,
        filename text NOT NULL,
        line integer NOT NULL
);
CREATE TABLE spawns (
        parentId text NOT NULL,
        childId text NOT NULL,
        filename text NOT NULL,
        line integer NOT NULL
);
CREATE TABLE time_events (
        timestamp integer NOT NULL,
        type text NOT NULL,
        id text NOT NULL,
        parentId text NOT NULL,
        childId text NOT NULL,
        filename text NOT NULL,
        line integer NOT NULL
);
CREATE UNIQUE INDEX idx_time_events_uniq_rows
ON time_events (type, id, parentId, childId, filename, line);

INSERT INTO time_events (timestamp, type, id, parentId, childId, filename, line)
VALUES (0, 'goroutine-start', 'main', '', '', '', 0);



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

CREATE VIEW goroutines_with_all_spawn_events AS
WITH RECURSIVE
        parents_count(childId, count) AS (
                SELECT spawns.childId, COUNT(*)
                FROM spawns
                GROUP BY spawns.childId
        ),
        has_all_spawn_events(id) AS (
                SELECT parents_count.childId AS id
                FROM parents_count
                INNER JOIN spawn_child_events_count ON spawn_child_events_count.childId = parents_count.childId
                WHERE spawn_child_events_count.count = parents_count.count
        ),
        spawn_child_events_count(id, count) AS (
                SELECT time_events.childId AS id, COUNT(*) AS count
                FROM time_events
                WHERE time_events.type = 'spawn-child'
                GROUP BY time_events.childId
        )
SELECT parents_count.childId AS id
FROM parents_count
INNER JOIN spawn_child_events_count ON parents_count.childId = spawn_child_events_count.id
WHERE parents_count.count = spawn_child_events_count.count;

CREATE VIEW new_spawn_child_events AS
SELECT  t4.maxTimestamp + (ROW_NUMBER() OVER (ORDER BY t1.parentId)) AS timestamp,
        'spawn-child' AS type,
        '' AS id,
        t1.parentId,
        t1.childId,
        t1.filename,
        t1.line
FROM spawns t1
CROSS JOIN (SELECT MAX(time_events.timestamp) AS maxTimestamp FROM time_events) AS t4
WHERE NOT EXISTS (
        SELECT 1
        FROM time_events t2
        WHERE   t2.type = 'spawn-child'
        AND     t2.id = ''
        AND     t1.parentId = t2.parentId
        AND     t1.childId = t2.childId
        AND     t1.filename = t2.filename
        AND     t1.line = t2.line)
AND EXISTS (
        SELECT 1
        FROM time_events t3
        WHERE   t3.id = t1.parentId
        AND     t3.type = 'goroutine-start');

CREATE VIEW new_goroutine_start_events AS
SELECT  t4.maxTimestamp + (ROW_NUMBER() OVER (ORDER BY t1.id)) AS timestamp,
        'goroutine-start' AS type,
        t1.id,
        '' AS parentId,
        '' AS childId,
        t1.filename,
        t1.line
FROM goroutines t1
INNER JOIN goroutines_with_all_spawn_events ON goroutines_with_all_spawn_events.id = t1.id
CROSS JOIN (SELECT MAX(time_events.timestamp) AS maxTimestamp FROM time_events) AS t4
WHERE NOT EXISTS (
        SELECT 1
        FROM time_events t2
        WHERE   t2.type = 'goroutine-start'
        AND     t2.id = t1.id
        AND     t2.parentId = ''
        AND     t2.childId = ''
        AND     t1.filename = t2.filename
        AND     t1.line = t2.line);

-- name: insert-spawn-child-events
INSERT INTO time_events SELECT * FROM new_spawn_child_events;

-- name: insert-spawn-child-events
INSERT INTO time_events SELECT * FROM new_goroutine_start_events;

-- name: insert-goroutine
INSERT INTO goroutines (id, packageName, filename, line)
VALUES (?, ?, ?, ?);

-- name: insert-spawn
INSERT INTO spawns (parentId, childId, filename, line)
VALUES (?, ?, ?, ?);
