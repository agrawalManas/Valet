CREATE TABLE drivers(
    id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    site_id UUID NOT NULL REFERENCES sites(id),
    created_by UUID NOT NULL REFERENCES users(id),

    -- Driver's shift timings
    shift_start_time TIME NOT NULL,
    shift_end_time TIME NOT NULL CHECK (shift_start_time <> shift_end_time),

    -- Driver's current status
    current_status VARCHAR(20) NOT NULL DEFAULT 'INACTIVE'
        CHECK (
            current_status IN (
                'AVAILABLE',
                'ON_DUTY',
                'OFF_DUTY',
                'SUSPENDED',
                'INACTIVE'
            )
        ),

    -- Login/logout tracking
    last_login TIMESTAMPTZ,
    last_logout TIMESTAMPTZ,

    -- Lifetime tips earned
    total_tip_earned NUMERIC(12, 2) NOT NULL DEFAULT 0
        CHECK (total_tip_earned >= 0),

    -- Documents
    aadhaar_number TEXT,
    aadhaar_url VARCHAR(255),
    pan_number VARCHAR(10),
    pan_url VARCHAR(255),
    driving_license VARCHAR(255),
    driving_license_url VARCHAR(255),

    -- KYC status
    kyc_status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (
            kyc_status IN (
                'PENDING',
                'SUBMITTED',
                'VERIFIED',
                'REJECTED'
            )
        ),

    -- Driving experience in completed years
    driving_experience_years SMALLINT NOT NULL CHECK (driving_experience_years >= 0),

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);