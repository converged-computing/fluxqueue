-- A single row for each group
CREATE TABLE pending_queue (
    jobspec TEXT NOT NULL,
    namespace TEXT NOT NULL,
    flux_job_name TEXT NOT NULL,
    name TEXT NOT NULL,
    type INTEGER NOT NULL,
    reservation INTEGER NOT NULL,
    duration INTEGER NOT NULL,
    created_at timestamptz NOT NULL default NOW(),    
    size INTEGER NOT NULL
);

CREATE UNIQUE INDEX pending_index ON pending_queue (name, namespace);

-- We only need the fluxid for a reservation
-- CREATE TABLE reservations (
--     group_name TEXT NOT NULL,
--     flux_id INTEGER NOT NULL
-- );