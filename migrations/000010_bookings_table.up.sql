CREATE TABLE IF NOT EXISTS bookings (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id),     -- the customer
    created_by        UUID NOT NULL REFERENCES users(id),     -- the customer or a supervisor
    site_id           UUID NOT NULL REFERENCES sites(id),
    slot_id           UUID REFERENCES slots(id),
    vehicle_number    VARCHAR(12) NOT NULL,                   -- normalized: upper case, letters and digits only

    for_date          DATE NOT NULL,                         
    status            VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED',
    booking_type      VARCHAR(10) NOT NULL,
    booking_mode      VARCHAR(10) NOT NULL,

    -- Amounts are copied from the site when the booking is made, so a later price
    -- change never alters an old booking. They are computed on the server.
    valet_amount      NUMERIC(10,2) NOT NULL,
    insurance_opted   BOOLEAN NOT NULL DEFAULT false,
    insurance_amount  NUMERIC(10,2) NOT NULL DEFAULT 0,
    taxes             NUMERIC(10,2) NOT NULL DEFAULT 0,
    total_amount      NUMERIC(12,2) GENERATED ALWAYS AS (valet_amount + insurance_amount + taxes) STORED,

    payment_method    VARCHAR(15),
    payment_status    VARCHAR(20) NOT NULL DEFAULT 'PENDING',

    cancelled_at      TIMESTAMPTZ,
    cancel_reason     TEXT,

    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_bookings_vehicle      CHECK (vehicle_number ~ '^[A-Z0-9]{6,12}$'),
    CONSTRAINT chk_bookings_status       CHECK (status IN ('SCHEDULED', 'ACTIVE', 'COMPLETED', 'CANCELLED', 'NO_SHOW')),
    CONSTRAINT chk_bookings_type         CHECK (booking_type IN ('PREPAID', 'GMC')),
    CONSTRAINT chk_bookings_mode         CHECK (booking_mode IN ('B2B', 'B2C')),
    CONSTRAINT chk_bookings_method       CHECK (payment_method IS NULL OR payment_method IN ('CASH', 'UPI', 'CARD', 'NET BANKING')),
    CONSTRAINT chk_bookings_pay_status   CHECK (payment_status IN ('NOT_REQUIRED', 'PENDING', 'PAID', 'FAILED', 'REFUNDED')),
    CONSTRAINT chk_bookings_amounts      CHECK (valet_amount >= 0 AND insurance_amount >= 0 AND taxes >= 0),
    CONSTRAINT chk_bookings_insurance    CHECK ((insurance_opted AND insurance_amount > 0)
                                             OR (NOT insurance_opted AND insurance_amount = 0)),
    -- PREPAID needs a payment method and a real payment status. GMC has neither.
    CONSTRAINT chk_bookings_payment      CHECK (
        (booking_type = 'PREPAID' AND payment_method IS NOT NULL AND payment_status <> 'NOT_REQUIRED')
     OR (booking_type = 'GMC'     AND payment_method IS NULL     AND payment_status = 'NOT_REQUIRED')),
    CONSTRAINT chk_bookings_slot         CHECK (status <> 'SCHEDULED' OR slot_id IS NOT NULL),
    CONSTRAINT chk_bookings_cancelled    CHECK ((status = 'CANCELLED') = (cancelled_at IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS idx_bookings_user         ON bookings (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bookings_site_day     ON bookings (site_id, for_date, status);
CREATE INDEX IF NOT EXISTS idx_bookings_site_vehicle ON bookings (site_id, vehicle_number);
CREATE INDEX IF NOT EXISTS idx_bookings_slot_day     ON bookings (slot_id, for_date);

-- A vehicle cannot be parked twice at the same time.
CREATE UNIQUE INDEX IF NOT EXISTS uq_bookings_one_active_per_vehicle
    ON bookings (vehicle_number) WHERE status = 'ACTIVE';

-- The same vehicle cannot hold the same slot on the same day twice.
CREATE UNIQUE INDEX IF NOT EXISTS uq_bookings_vehicle_slot_day
    ON bookings (vehicle_number, for_date, slot_id)
    WHERE status IN ('SCHEDULED', 'ACTIVE') AND slot_id IS NOT NULL;