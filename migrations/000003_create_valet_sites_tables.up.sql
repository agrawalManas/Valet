CREATE TABLE IF NOT EXISTS sites (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_name                VARCHAR(255) NOT NULL,
    is_booking_allowed       BOOLEAN NOT NULL,
    parking_type             VARCHAR(20) NOT NULL,
    total_capacity           INTEGER NOT NULL,
    driver_count             INTEGER NOT NULL DEFAULT 0,
    site_opening_time        TIMESTAMPTZ NOT NULL,
    site_closing_time        TIMESTAMPTZ NOT NULL,
    valet_fee                NUMERIC(10,2) NOT NULL DEFAULT 0,
    site_images              TEXT[],

    -- SiteAddress
    pin_code                 VARCHAR(6) NOT NULL,
    city                     VARCHAR(100) NOT NULL,
    state                    VARCHAR(100) NOT NULL,
    address_line_1           VARCHAR(255) NOT NULL,
    address_line_2           VARCHAR(255),
    site_coordinates         TEXT,

    -- Supervisor
    supervisor_id            UUID NOT NULL REFERENCES users(id),

    -- SiteInsuranceAddOn
    insurance_add_on_allowed BOOLEAN NOT NULL,
    insurance_add_on_price   NUMERIC(10,2),

    -- TippingDetails
    tipping_allowed          BOOLEAN NOT NULL,
    tipping_amount           NUMERIC(10,2)[],
    pick_tipping_allowed     BOOLEAN,
    drop_tipping_allowed     BOOLEAN,

    created_by               UUID NOT NULL REFERENCES users(id),
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_sites_parking_type CHECK (parking_type IN ('BASEMENT', 'GROUND', 'OTHER')),
    CONSTRAINT chk_sites_capacity     CHECK (total_capacity >= 1),
    CONSTRAINT chk_sites_drivers      CHECK (driver_count >= 0),
    CONSTRAINT chk_sites_valet_fee    CHECK (valet_fee >= 0),
    CONSTRAINT chk_sites_hours        CHECK (site_opening_time < site_closing_time),
    CONSTRAINT chk_sites_pin_code     CHECK (pin_code ~ '^[0-9]{6}$'),
    CONSTRAINT chk_sites_insurance    CHECK (
        (NOT insurance_add_on_allowed OR insurance_add_on_price IS NOT NULL)
        AND (insurance_add_on_price IS NULL OR insurance_add_on_price > 0)
    ),
    CONSTRAINT chk_sites_tipping      CHECK (
        (NOT tipping_allowed OR tipping_amount IS NOT NULL)
        AND (tipping_amount IS NULL
             OR (cardinality(tipping_amount) BETWEEN 1 AND 10 AND 0 < ALL (tipping_amount)))
    )
);

CREATE INDEX IF NOT EXISTS idx_sites_supervisor ON sites (supervisor_id);
CREATE INDEX IF NOT EXISTS idx_sites_created_by ON sites (created_by);