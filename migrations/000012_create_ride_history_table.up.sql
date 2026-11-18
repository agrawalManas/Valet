CREATE TABLE ride_history (
    sno SERIAL PRIMARY KEY,
    ride_id UUID NOT NULL REFERENCES rides(id),
    from_status VARCHAR(20),                          -- NULL on the first row
    to_status   VARCHAR(20) NOT NULL,
    actor_id    UUID REFERENCES users(id),            -- NULL when the system acted (auto-complete)
    actor_role  VARCHAR(20),                          -- role at that moment
    driver_id   UUID REFERENCES drivers(id),          -- set on assignment events
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP

    CONSTRAINT chk_ride_history_from CHECK (from_status IS NULL OR from_status IN (
        'REQUESTED', 'PICK_ASSIGNED', 'PICK_COMPLETED', 'PARKED', 'GMC_REQUESTED', 'DROP_ASSIGNED',
        'DRIVER_ENROUTE', 'DROP_COMPLETED', 'KEY_RETURNED', 'RIDE_FEEDBACK', 'COMPLETED', 'CANCELLED')),
    CONSTRAINT chk_ride_history_to CHECK (to_status IN (
        'REQUESTED', 'PICK_ASSIGNED', 'PICK_COMPLETED', 'PARKED', 'GMC_REQUESTED', 'DROP_ASSIGNED',
        'DRIVER_ENROUTE', 'DROP_COMPLETED', 'KEY_RETURNED', 'RIDE_FEEDBACK', 'COMPLETED', 'CANCELLED'))
);