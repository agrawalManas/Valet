-- 1. Add new column
ALTER TABLE drivers ADD COLUMN current_supervisor UUID;

-- 2. Backfill from existing column
UPDATE drivers SET current_supervisor = created_by;

-- 3. Enforce NOT NULL on the new column
ALTER TABLE drivers ALTER COLUMN current_supervisor SET NOT NULL;   