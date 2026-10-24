CREATE TABLE IF NOT EXISTS user_access_audit (
    id             SERIAL PRIMARY KEY,
    user_id        UUID NOT NULL REFERENCES users(id),   -- whose access changed
    changed_by     UUID NOT NULL REFERENCES users(id),   -- the admin who did it
    old_role       VARCHAR(255),
    new_role       VARCHAR(255),
    old_is_active  BOOLEAN,
    new_is_active  BOOLEAN,
    reason         TEXT,                                  -- optional note
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_audit_old_role CHECK (old_role IN ('CUSTOMER','DRIVER','SUPERVISOR','ADMIN')),
    CONSTRAINT chk_audit_new_role CHECK (new_role IN ('CUSTOMER','DRIVER','SUPERVISOR','ADMIN'))
);

CREATE INDEX idx_user_access_audit_user    ON user_access_audit (user_id, created_at DESC);
CREATE INDEX idx_user_access_audit_actor   ON user_access_audit (changed_by, created_at DESC);