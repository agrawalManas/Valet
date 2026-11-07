CREATE TABLE driver_supervisors (
    driver_id UUID NOT NULL REFERENCES users(id),
    supervisor_id UUID NOT NULL REFERENCES users(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);