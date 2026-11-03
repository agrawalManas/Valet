CREATE TABLE driver_supervisors (
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    supervisor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (driver_id, supervisor_id)
);