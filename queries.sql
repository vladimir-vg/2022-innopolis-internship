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

CREATE VIEW time_events AS
WITH RECURSIVE
        time_events0(timestamp, type, id, parentId, childId) AS (
                SELECT 0 AS timestamp, 'goroutine-start' AS type, 'main' AS id, NULL AS parentId, NULL AS childId
                UNION ALL
                SELECT
                        (spawn_child_event.maxTimestamp + spawn_child_event.rowNumber) AS timestamp,
                        'spawn-child' AS type, NULL as id, spawn_child_event.parentId, spawn_child_event.childId
                FROM spawn_child_event
        ),
        spawn_child_event(maxTimestamp, rowNumber, parentId, childId) AS (
                SELECT
                        parent_start_event_max_timestamp.maxTimestamp,
                        ROW_NUMBER() OVER (PARTITION BY spawns.parentId) AS rowNumber,
                        spawns.parentId, spawns.childId
                FROM spawns
                INNER JOIN all_parents_have_start_events ON all_parents_have_start_events.childId = spawns.childId
                INNER JOIN parent_start_event_max_timestamp ON parent_start_event_max_timestamp.childId = spawns.childId
        ),
        all_parents_have_start_events(childId) AS (
                SELECT parent_start_events_count.childId
                FROM parent_start_events_count
                INNER JOIN parents_count ON parent_start_events_count.childId = parents_count.childId
                WHERE parent_start_events_count.count = parents_count.count
        ),
        parent_start_events_count(childId, count) AS (
                SELECT spawns.childId, COUNT(*) AS count
                FROM spawns
                INNER JOIN time_events0 ON time_events0.parentId = spawns.parentId
                WHERE time_events0.type = 'goroutine-start'
                GROUP BY spawns.childId
        ),
        parents_count(childId, count) AS (
                SELECT spawns.childId, COUNT(*)
                FROM spawns
                GROUP BY spawns.childId
        ),
        parent_start_event_max_timestamp(childId, maxTimestamp) AS (
                SELECT spawns.childId, MAX(time_events0.timestamp) AS timestamp
                FROM spawns
                INNER JOIN time_events0 ON time_events0.parentId = spawns.parentId
                WHERE time_events0.type = 'goroutine-start'
                GROUP BY spawns.childId
        )
SELECT * FROM time_events0;

-- name: insert-goroutine
INSERT INTO goroutines (id, packageName, filename, line)
VALUES (?, ?, ?, ?);

-- name: insert-spawn
INSERT INTO spawns (parentId, childId, filename, line)
VALUES (?, ?, ?, ?);
